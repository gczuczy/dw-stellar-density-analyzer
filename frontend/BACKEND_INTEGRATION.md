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

**Example Implementation (Node.js/Express):**
```javascript
app.get('/api/oauth/config', (req, res) => {
  res.json({
    issuer: process.env.OAUTH_ISSUER,
    clientId: process.env.OAUTH_CLIENT_ID,
    redirectUri: `${req.protocol}://${req.get('host')}/api/auth/callback`,
    scope: 'openid profile email'
  });
});
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

**Example Implementation (Node.js/Express with Passport.js):**

```javascript
const passport = require('passport');
const { Strategy: OidcStrategy } = require('passport-openidconnect');

// Configure OIDC strategy
passport.use('oidc', new OidcStrategy({
  issuer: process.env.OAUTH_ISSUER,
  authorizationURL: process.env.OAUTH_AUTH_URL,
  tokenURL: process.env.OAUTH_TOKEN_URL,
  userInfoURL: process.env.OAUTH_USERINFO_URL,
  clientID: process.env.OAUTH_CLIENT_ID,
  clientSecret: process.env.OAUTH_CLIENT_SECRET,
  callbackURL: '/api/auth/callback',
  scope: 'openid profile email'
}, (issuer, profile, done) => {
  // Store user profile in session
  return done(null, {
    id: profile.id,
    email: profile.emails[0].value,
    name: profile.displayName,
    claims: profile._json
  });
}));

// Serialize user to session
passport.serializeUser((user, done) => {
  done(null, user);
});

passport.deserializeUser((user, done) => {
  done(null, user);
});

// Callback route
app.get('/api/auth/callback',
  passport.authenticate('oidc', { failureRedirect: '/login-failed' }),
  (req, res) => {
    // Authentication successful
    // Redirect back to frontend with the original code and state
    const { code, state } = req.query;
    res.redirect(`/?code=${code}&state=${state}`);
  }
);
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

**Example Implementation:**
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

**Example Implementation:**
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

```javascript
app.use(cors({
  origin: 'https://your-frontend-domain.com',
  credentials: true
}));
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

```bash
# OAuth2/OIDC Provider
OAUTH_ISSUER=https://your-idp.example.com
OAUTH_CLIENT_ID=your-spa-client-id
OAUTH_CLIENT_SECRET=your-client-secret  # Required for code exchange
OAUTH_AUTH_URL=https://your-idp.example.com/authorize
OAUTH_TOKEN_URL=https://your-idp.example.com/oauth/token
OAUTH_USERINFO_URL=https://your-idp.example.com/userinfo

# Session
SESSION_SECRET=your-random-secret-key-here
SESSION_COOKIE_NAME=edsurvey.sid

# Application
PORT=8081
NODE_ENV=development
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
