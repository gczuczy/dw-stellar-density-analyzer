# Backend Integration Guide

This document describes the backend endpoints required for the ED Survey Tools frontend to function properly.

## Required Backend Endpoints

### 1. OAuth Configuration Endpoint

**Endpoint:** `GET /api/oauth/config`

**Purpose:** Provides OAuth2/OIDC configuration to the frontend.

**Response Format:**
```json
{
  "issuer": "https://your-idp.example.com",
  "clientId": "your-spa-client-id",
  "redirectUri": "http://localhost:4200/api/auth/callback",
  "scope": "openid profile email"
}
```

**Fields:**
- `issuer` (required) - The OAuth2/OIDC provider's issuer URL. The frontend will fetch `/.well-known/openid-configuration` from this URL.
- `clientId` (required) - The client ID registered at the IdP for this SPA.
- `redirectUri` (optional) - The callback URL. Defaults to `{origin}/api/auth/callback` if not provided.
- `scope` (optional) - OAuth2 scopes to request. Defaults to `"openid profile email"`.

**Example Implementation:**
```bash
app.get('/api/oauth/config', (req, res) => {
  res.json({
    issuer: process.env.OAUTH_ISSUER,
    clientId: process.env.OAUTH_CLIENT_ID,
    redirectUri: `${req.protocol}://${req.get('host')}/api/auth/callback`,
    scope: 'openid profile email'
  });
});
```

**Or in Go:**
```go
func handleOAuthConfig(w http.ResponseWriter, r *http.Request) {
  config := map[string]string{
    "issuer":      os.Getenv("OAUTH_ISSUER"),
    "clientId":    os.Getenv("OAUTH_CLIENT_ID"),
    "redirectUri": "http://localhost:4200/api/auth/callback",
    "scope":       "openid profile email",
  }
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(config)
}
```

---

### 2. OAuth Callback Endpoint (CRITICAL)

**Endpoint:** `GET /api/auth/callback`

**Purpose:** Handles the OAuth2 authorization code callback from the IdP, exchanges the code for tokens, stores user session, and redirects back to the frontend.

**Flow:**
```
1. IdP redirects here with: ?code=xxx&state=yyy
2. Backend exchanges code for tokens with IdP
3. Backend extracts user identity (sub, email, etc.)
4. Backend creates server-side session
5. Backend redirects to frontend with session cookie
```

#### Implementation Requirements

**Query Parameters (from IdP):**
- `code` - Authorization code from the IdP
- `state` - State parameter (PKCE code challenge is handled by angular-oauth2-oidc)

**Implementation Steps:**

1. **Exchange authorization code for tokens**
   - Make POST request to IdP token endpoint
   - Include: code, client_id, client_secret (if applicable), redirect_uri, grant_type=authorization_code
   - Receive: access_token, id_token, refresh_token (optional)

2. **Validate and decode ID token**
   - Verify signature using IdP's JWKS
   - Validate issuer, audience, expiration
   - Extract user claims (sub, email, name, etc.)

3. **Create server-side session**
   - Store user identity in session store (Redis, database, etc.)
   - Set secure session cookie
   - Store tokens securely (encrypted, HttpOnly cookie or server-side only)

4. **Redirect back to frontend**
   - Redirect to: `/?code={authorization_code}&state={state}`
   - The frontend's angular-oauth2-oidc will complete the PKCE flow
   - Frontend will then have the access token in browser sessionStorage

**Example Implementation (Go with Gorilla sessions and go-oidc):**

```go
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

var (
	oauth2Config *oauth2.Config
	oidcVerifier *oidc.IDTokenVerifier
	store        *sessions.CookieStore
)

func init() {
	// Initialize session store
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = generateRandomString(32)
		log.Println("Warning: Using random session secret")
	}
	store = sessions.NewCookieStore([]byte(sessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   os.Getenv("NODE_ENV") == "production",
		SameSite: http.SameSiteLaxMode,
	}
}

func main() {
	ctx := context.Background()

	// Initialize OIDC provider
	provider, err := oidc.NewProvider(ctx, os.Getenv("OAUTH_ISSUER"))
	if err != nil {
		log.Fatal(err)
	}

	// Configure OAuth2
	oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:4200/api/auth/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// Configure OIDC verifier
	oidcVerifier = provider.Verifier(&oidc.Config{
		ClientID: os.Getenv("OAUTH_CLIENT_ID"),
	})

	// Routes
	http.HandleFunc("/api/oauth/config", handleOAuthConfig)
	http.HandleFunc("/api/auth/callback", handleAuthCallback)
	http.HandleFunc("/api/auth/user", handleAuthUser)
	http.HandleFunc("/api/auth/logout", handleLogout)

	log.Println("Server starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

// OAuth configuration endpoint
func handleOAuthConfig(w http.ResponseWriter, r *http.Request) {
	config := map[string]string{
		"issuer":      os.Getenv("OAUTH_ISSUER"),
		"clientId":    os.Getenv("OAUTH_CLIENT_ID"),
		"redirectUri": "http://localhost:4200/api/auth/callback",
		"scope":       "openid profile email",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// OAuth callback handler
func handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authorization code from query
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for tokens
	oauth2Token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		log.Printf("Code exchange failed: %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Extract ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token in token response", http.StatusInternalServerError)
		return
	}

	// Verify ID token
	idToken, err := oidcVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Printf("ID token verification failed: %v", err)
		http.Error(w, "Failed to verify ID token", http.StatusInternalServerError)
		return
	}

	// Extract claims
	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "Failed to parse claims", http.StatusInternalServerError)
		return
	}

	// Create session
	session, _ := store.Get(r, "auth-session")
	session.Values["user_id"] = claims.Sub
	session.Values["user_email"] = claims.Email
	session.Values["user_name"] = claims.Name
	session.Values["access_token"] = oauth2Token.AccessToken

	if err := session.Save(r, w); err != nil {
		log.Printf("Failed to save session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Redirect back to frontend with original code and state
	// This allows the frontend to complete its PKCE flow
	redirectURL := "/?code=" + code
	if state != "" {
		redirectURL += "&state=" + state
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// User info endpoint
func handleAuthUser(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth-session")

	userID, ok := session.Values["user_id"].(string)
	if !ok || userID == "" {
		json.NewEncoder(w).Encode(map[string]bool{
			"authenticated": false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"userId":        session.Values["user_id"],
		"email":         session.Values["user_email"],
		"name":          session.Values["user_name"],
		"authenticated": true,
	})
}

// Logout endpoint
func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, _ := store.Get(r, "auth-session")
	session.Options.MaxAge = -1 // Delete session

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

// Utility function to generate random strings
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
```

**Required Go dependencies:**
```bash
go get github.com/coreos/go-oidc/v3/oidc
go get github.com/gorilla/sessions
go get golang.org/x/oauth2
```

**go.mod example:**
```go
module your-backend

go 1.21

require (
	github.com/coreos/go-oidc/v3 v3.9.0
	github.com/gorilla/sessions v1.2.2
	golang.org/x/oauth2 v0.16.0
)
```

**Alternative: Simpler Implementation (Python/Flask)**

```python
from flask import Flask, redirect, request, session
import requests
import jwt

app = Flask(__name__)
app.secret_key = 'your-secret-key'

@app.route('/api/auth/callback')
def auth_callback():
    # Get authorization code from IdP
    code = request.args.get('code')
    state = request.args.get('state')
    
    if not code:
        return 'Missing authorization code', 400
    
    # Exchange code for tokens
    token_response = requests.post(
        f"{OAUTH_ISSUER}/oauth/token",
        data={
            'grant_type': 'authorization_code',
            'code': code,
            'redirect_uri': f"{request.host_url}api/auth/callback",
            'client_id': OAUTH_CLIENT_ID,
            'client_secret': OAUTH_CLIENT_SECRET
        }
    )
    
    if token_response.status_code != 200:
        return 'Token exchange failed', 500
    
    tokens = token_response.json()
    
    # Decode ID token to get user info
    id_token = tokens['id_token']
    user_info = jwt.decode(id_token, options={"verify_signature": False})
    
    # Store user in session
    session['user_id'] = user_info['sub']
    session['user_email'] = user_info.get('email')
    session['user_name'] = user_info.get('name')
    session['access_token'] = tokens['access_token']
    
    # Redirect back to frontend with original code/state
    # The frontend will complete its own PKCE flow
    return redirect(f"/?code={code}&state={state}")
```

---

### 3. User Info Endpoint (Optional but Recommended)

**Endpoint:** `GET /api/auth/user`

**Purpose:** Returns the currently authenticated user's information from the server-side session.

**Response Format:**
```json
{
  "userId": "user-123-from-idp",
  "email": "user@example.com",
  "name": "John Doe",
  "authenticated": true
}
```

**When Not Authenticated:**
```json
{
  "authenticated": false
}
```

**Example Implementation (Node.js):**
```javascript
app.get('/api/auth/user', (req, res) => {
  if (req.user) {
    res.json({
      userId: req.user.id,
      email: req.user.email,
      name: req.user.name,
      authenticated: true
    });
  } else {
    res.json({ authenticated: false });
  }
});
```

**Example Implementation (Go):**
```go
func handleAuthUser(w http.ResponseWriter, r *http.Request) {
  session, _ := store.Get(r, "auth-session")
  
  userID, ok := session.Values["user_id"].(string)
  if !ok || userID == "" {
    json.NewEncoder(w).Encode(map[string]bool{
      "authenticated": false,
    })
    return
  }
  
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(map[string]interface{}{
    "userId":        session.Values["user_id"],
    "email":         session.Values["user_email"],
    "name":          session.Values["user_name"],
    "authenticated": true,
  })
}
```

---

### 4. Logout Endpoint (Optional)

**Endpoint:** `POST /api/auth/logout`

**Purpose:** Destroys the server-side session.

**Response:**
```json
{
  "success": true
}
```

**Example Implementation (Node.js):**
```javascript
app.post('/api/auth/logout', (req, res) => {
  req.session.destroy((err) => {
    if (err) {
      return res.status(500).json({ success: false });
    }
    res.clearCookie('connect.sid'); // or your session cookie name
    res.json({ success: true });
  });
});
```

**Example Implementation (Go):**
```go
func handleLogout(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodPost {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }
  
  session, _ := store.Get(r, "auth-session")
  session.Options.MaxAge = -1 // Delete session
  
  if err := session.Save(r, w); err != nil {
    http.Error(w, "Failed to logout", http.StatusInternalServerError)
    return
  }
  
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
```

---

## Security Considerations

### Session Management
- Use secure, HttpOnly cookies for session tokens
- Set `SameSite=Lax` or `SameSite=Strict` on cookies
- Use HTTPS in production (required for secure cookies)
- Implement CSRF protection

### Token Storage
- **DO NOT** store access tokens in frontend localStorage
- **DO** store tokens server-side only
- **DO** use short-lived access tokens (15-60 minutes)
- **DO** use refresh tokens for long-lived sessions

### CORS Configuration
Since frontend and backend are on the same origin (proxied via `/api`), CORS is not an issue in development. In production:

**Node.js/Express:**
```javascript
app.use(cors({
  origin: 'https://your-frontend-domain.com',
  credentials: true
}));
```

**Go with rs/cors:**
```go
import "github.com/rs/cors"

c := cors.New(cors.Options{
  AllowedOrigins:   []string{"https://your-frontend-domain.com"},
  AllowCredentials: true,
  AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
  AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
})

handler := c.Handler(http.DefaultServeMux)
http.ListenAndServe(":8081", handler)
```

---

## OAuth2 Flow Diagram

```
┌─────────┐                                    ┌─────────┐
│ Browser │                                    │   IdP   │
│(Angular)│                                    │         │
└────┬────┘                                    └────┬────┘
     │                                              │
     │ 1. Click Login                               │
     │────────────────────────────────────────────► │
     │ GET /authorize?client_id=...&redirect_uri=   │
     │     /api/auth/callback&code_challenge=...    │
     │                                              │
     │ 2. User authenticates                        │
     │ ◄────────────────────────────────────────────│
     │                                              │
     │ 3. Redirect with code                        │
┌────▼────┐                                         │
│ Backend │ ◄───────────────────────────────────────┘
│         │ GET /api/auth/callback?code=xxx&state=yyy
└────┬────┘
     │ 4. Exchange code for tokens
     │────────────────────────────────────────────►
     │ POST /token {code, client_secret, ...}      │
     │                                             │
     │ 5. Receive tokens                           │
     │ ◄────────────────────────────────────────────
     │ {access_token, id_token, refresh_token}     │
     │                                             │
     │ 6. Create session, set cookie                │
     │                                              │
     │ 7. Redirect to frontend                      │
     │────────────────────────────────────────────►
     │ 302 /?code=xxx&state=yyy                  ┌─────────┐
     │                                           │ Browser │
     │                                           │(Angular)│
     │                                           └────┬────┘
     │                                                │
     │ 8. Frontend completes PKCE flow                │
     │ ◄──────────────────────────────────────────────┘
     │ (angular-oauth2-oidc handles this)             │
     │                                                │
     │ 9. Both frontend and backend have user session │
     └────────────────────────────────────────────────┘
```

---

## Testing the Integration

### 1. Start Backend
```bash
# Your backend should be running on port 8081
npm start  # or python app.py, etc.
```

### 2. Start Frontend
```bash
cd ed-survey-tools
npm start
# Opens http://localhost:4200
# All /api/* requests proxy to http://localhost:8081
```

### 3. Test Authentication Flow
1. Open http://localhost:4200
2. Click "Login"
3. Redirected to IdP
4. Authenticate
5. Redirected to `/api/auth/callback` (backend)
6. Backend processes and redirects to `/?code=...`
7. Frontend completes PKCE and shows logged-in state

### 4. Verify Server-Side Session
```bash
# Make authenticated request to backend
curl -b cookies.txt http://localhost:8081/api/auth/user
```

---

## Environment Variables

Your backend should configure these environment variables:

**For Go backend:**
```bash
# OAuth2/OIDC Provider
OAUTH_ISSUER=https://your-idp.example.com
OAUTH_CLIENT_ID=your-spa-client-id
OAUTH_CLIENT_SECRET=your-client-secret  # Required for code exchange

# Session
SESSION_SECRET=your-random-secret-key-here

# Application
PORT=8081
NODE_ENV=development  # or production
```

**For Python backend:**
```bash
# Same as above, plus:
OAUTH_AUTH_URL=https://your-idp.example.com/authorize
OAUTH_TOKEN_URL=https://your-idp.example.com/oauth/token
OAUTH_USERINFO_URL=https://your-idp.example.com/userinfo
SESSION_COOKIE_NAME=edsurvey.sid
```

---

## IdP Configuration

Register your application at the IdP with these settings:

**Redirect URIs:**
- Development: `http://localhost:4200/api/auth/callback`
- Production: `https://yourdomain.com/api/auth/callback`

**Grant Types:**
- Authorization Code
- Refresh Token (optional)

**Client Authentication:**
- Confidential client (use client secret)
- PKCE enabled (optional but recommended)

**Scopes:**
- `openid` (required)
- `profile`
- `email`
- Any custom scopes your app needs

---

## Troubleshooting

### "Invalid redirect_uri"
- Ensure the redirect URI in `/api/oauth/config` matches exactly what's registered at the IdP
- Check for trailing slashes, http vs https, port numbers

### "Code exchange failed"
- Verify client_secret is correct
- Check that redirect_uri in token request matches the one used in authorization request
- Ensure code hasn't expired (typically 60 seconds)

### "Session not created"
- Check session middleware is configured correctly
- Verify session secret is set
- Ensure cookies are being sent (credentials: true in CORS)

### "Frontend can't complete PKCE"
- The frontend MUST receive the original `code` and `state` parameters
- Backend redirect must be: `/?code={code}&state={state}`
- Don't consume or modify these parameters

---

## Summary

**What the backend MUST do:**

1. ✅ Serve OAuth config at `/api/oauth/config`
2. ✅ Handle callback at `/api/auth/callback`
3. ✅ Exchange authorization code for tokens
4. ✅ Extract user identity from tokens
5. ✅ Create server-side session
6. ✅ Redirect back to frontend with code/state preserved

**What the backend gets:**

- User identity (sub, email, name, etc.) from ID token
- Server-side session with authenticated user
- Ability to make API calls on behalf of the user

**What the frontend gets:**

- Access token (via angular-oauth2-oidc PKCE flow)
- User info from ID token claims
- Automatic token refresh (if refresh tokens are supported)
