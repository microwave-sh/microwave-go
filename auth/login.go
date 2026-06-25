package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// LoginMode selects which grant the interactive login uses.
type LoginMode int

const (
	// LoginAuto tries the loopback authorization-code+PKCE flow first and falls
	// back to the device grant when no browser/loopback is available. Default.
	LoginAuto LoginMode = iota
	// LoginLoopback forces the authorization-code+PKCE loopback flow.
	LoginLoopback
	// LoginDevice forces the device-authorization grant.
	LoginDevice
)

// LoginConfig drives Login. MetadataURL + ClientID are the only required
// fields; everything else has a sensible default. Consumers (CLIs) discover
// MetadataURL + ClientID from their product's public auth-config document.
type LoginConfig struct {
	// MetadataURL is the RFC 8414 authorization-server metadata document.
	MetadataURL string
	// ClientID identifies the Microwave key spec / trust exchange this login
	// mints against.
	ClientID string
	// Scopes is the optional requested scope set.
	Scopes []string
	// Mode selects the grant; zero value is LoginAuto.
	Mode LoginMode
	// HTTPClient overrides the default client used for all back-channel calls.
	HTTPClient *http.Client
	// Store, when set, persists the minted credentials.
	Store TokenStore
	// OpenBrowser overrides how the verification/authorize URL is opened.
	// Return an error to signal "no browser" (Login falls back to device in
	// LoginAuto). Defaults to the OS handler.
	OpenBrowser func(string) error
	// Output receives human-facing instructions (the device user code / URL).
	// Defaults to os.Stderr via the caller; nil discards.
	Output io.Writer
}

// errNoBrowser signals that the loopback flow can't proceed interactively, so
// LoginAuto should fall back to the device grant.
var errNoBrowser = errors.New("microwave/auth: no browser available")

// Login performs an interactive login and returns ready-to-use credentials,
// persisting them when a Store is configured. In LoginAuto it uses the
// loopback authorization-code+PKCE flow, falling back to the device grant.
func Login(ctx context.Context, cfg LoginConfig) (*Credentials, error) {
	if strings.TrimSpace(cfg.MetadataURL) == "" {
		return nil, fmt.Errorf("microwave/auth: MetadataURL is required")
	}
	if strings.TrimSpace(cfg.ClientID) == "" {
		return nil, fmt.Errorf("microwave/auth: ClientID is required")
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = defaultHTTPClient()
	}

	md, err := fetchMetadata(ctx, httpClient, cfg.MetadataURL)
	if err != nil {
		return nil, err
	}

	creds, err := runLogin(ctx, cfg, md, httpClient)
	if err != nil {
		return nil, err
	}
	if cfg.Store != nil {
		if err := cfg.Store.Save(creds); err != nil {
			return nil, err
		}
	}
	return creds, nil
}

func runLogin(ctx context.Context, cfg LoginConfig, md *ASMetadata, httpClient *http.Client) (*Credentials, error) {
	switch cfg.Mode {
	case LoginDevice:
		if !md.supportsDeviceGrant() {
			return nil, fmt.Errorf("microwave/auth: server does not advertise a device authorization endpoint")
		}
		return loginDevice(ctx, cfg, md, httpClient)
	case LoginLoopback:
		if md.AuthorizationEndpoint == "" {
			return nil, fmt.Errorf("microwave/auth: server does not advertise an authorization endpoint")
		}
		return loginLoopback(ctx, cfg, md, httpClient)
	default: // LoginAuto
		if md.AuthorizationEndpoint != "" {
			creds, err := loginLoopback(ctx, cfg, md, httpClient)
			if err == nil {
				return creds, nil
			}
			if !errors.Is(err, errNoBrowser) || !md.supportsDeviceGrant() {
				return nil, err
			}
			// Fall through to device grant.
		}
		if md.supportsDeviceGrant() {
			return loginDevice(ctx, cfg, md, httpClient)
		}
		return nil, fmt.Errorf("microwave/auth: server advertises neither an authorization nor a device endpoint")
	}
}

// callbackResult carries what the loopback handler captured from the redirect.
type callbackResult struct {
	code  string
	state string
	iss   string
	err   string
	desc  string
}

// loginLoopback runs RFC 8252 §7.3 loopback authorization-code + PKCE (RFC 7636).
func loginLoopback(ctx context.Context, cfg LoginConfig, md *ASMetadata, httpClient *http.Client) (*Credentials, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		// Can't bind a loopback port — let LoginAuto fall back to device.
		return nil, fmt.Errorf("%w: loopback bind failed: %v", errNoBrowser, err)
	}
	defer ln.Close()

	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", ln.Addr().(*net.TCPAddr).Port)

	pkce, err := newPKCE()
	if err != nil {
		return nil, err
	}
	state, err := randomURLToken(24)
	if err != nil {
		return nil, err
	}

	results := make(chan callbackResult, 1)
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/callback" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		res := callbackResult{
			code:  q.Get("code"),
			state: q.Get("state"),
			iss:   q.Get("iss"),
			err:   q.Get("error"),
			desc:  q.Get("error_description"),
		}
		if res.err != "" {
			writeBrowserMessage(w, "Login failed", "You can close this window and return to the terminal.")
		} else {
			writeBrowserMessage(w, "Login complete", "You're signed in. You can close this window.")
		}
		select {
		case results <- res:
		default:
		}
	})}
	go srv.Serve(ln) //nolint:errcheck
	defer srv.Close()

	authURL := buildAuthorizeURL(md.AuthorizationEndpoint, cfg.ClientID, redirectURI, pkce.challenge, state, cfg.Scopes)
	opener := cfg.OpenBrowser
	if opener == nil {
		opener = openBrowser
	}
	if err := opener(authURL); err != nil {
		return nil, fmt.Errorf("%w: %v", errNoBrowser, err)
	}
	fmt.Fprintf(output(cfg), "\n  If your browser didn't open, visit:\n  %s\n\n", authURL)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-results:
		if res.err != "" {
			return nil, &OAuthError{Code: res.err, Description: res.desc}
		}
		if res.state != state {
			return nil, fmt.Errorf("microwave/auth: state mismatch (possible CSRF); aborting")
		}
		// RFC 9207: when the AS advertises iss support it MUST be returned and
		// MUST match the metadata issuer.
		if md.IssParameterSupported && res.iss != md.Issuer {
			return nil, fmt.Errorf("microwave/auth: issuer mismatch in authorization response (RFC 9207)")
		}
		if res.code == "" {
			return nil, fmt.Errorf("microwave/auth: authorization response carried no code")
		}
		tok, err := postToken(ctx, httpClient, md.TokenEndpoint, url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {res.code},
			"redirect_uri":  {redirectURI},
			"client_id":     {cfg.ClientID},
			"code_verifier": {pkce.verifier},
		})
		if err != nil {
			return nil, err
		}
		return tok.credentials(md.TokenEndpoint, cfg.ClientID, time.Now()), nil
	}
}

func buildAuthorizeURL(endpoint, clientID, redirectURI, challenge, state string, scopes []string) string {
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"state":                 {state},
	}
	if len(scopes) > 0 {
		q.Set("scope", strings.Join(scopes, " "))
	}
	sep := "?"
	if strings.Contains(endpoint, "?") {
		sep = "&"
	}
	return endpoint + sep + q.Encode()
}

func output(cfg LoginConfig) io.Writer {
	if cfg.Output == nil {
		return io.Discard
	}
	return cfg.Output
}

func writeBrowserMessage(w http.ResponseWriter, title, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<!doctype html><meta charset=utf-8><title>%s</title><body style=\"font:16px system-ui;margin:3rem\"><h1>%s</h1><p>%s</p>", title, title, body)
}

func defaultHTTPClient() *http.Client { return &http.Client{Timeout: 30 * time.Second} }

// openBrowser opens url in the OS default browser (the RFC 8252 §5 external
// user-agent). Returns errNoBrowser when no opener is available.
func openBrowser(target string) error {
	var name string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		name, args = "open", []string{target}
	case "windows":
		name, args = "rundll32", []string{"url.dll,FileProtocolHandler", target}
	default:
		name, args = "xdg-open", []string{target}
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%w: %s not found", errNoBrowser, name)
	}
	if err := exec.Command(path, args...).Start(); err != nil {
		return fmt.Errorf("%w: %v", errNoBrowser, err)
	}
	return nil
}
