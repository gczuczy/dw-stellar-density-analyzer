import { Injectable }                       from '@angular/core';
import { OAuthService, AuthConfig }        from 'angular-oauth2-oidc';
import { Observable, from, of }           from 'rxjs';
import { map, switchMap, tap, catchError } from 'rxjs/operators';
import { Router }                         from '@angular/router';
import { ApiService }                     from '../services/api.service';

/**
 * The flat key→value shape that `/api/auth/config` must return.
 *
 * Because the primary OAuth2 provider does not expose
 * `.well-known/openid-configuration`, the backend supplies the
 * authorization and token endpoint URLs directly.
 */
export interface OAuthBackendConfig {
  clientId:            string;
  redirectUri:         string;
  scope:               string;
  issuer:              string;
  authUrl:             string;
  tokenUrl:            string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private configLoadFailed = false;

  constructor(
    private oauth  : OAuthService,
    private api    : ApiService,
    private router : Router
  ) {}

  /**
   * Called once by APP_INITIALIZER.
   *
   * If the OAuth config fails to load, we mark the service as disabled
   * and allow the app to continue loading (public pages will work).
   */
  initialize(): Observable<void> {
    return this.api.getConfig<OAuthBackendConfig>('/api/auth/config').pipe(

      tap((cfg: OAuthBackendConfig) => {
        // Validate required fields
        if (!cfg.clientId || !cfg.authUrl || !cfg.tokenUrl) {
          throw new Error('OAuth config missing required fields: clientId, authUrl, tokenUrl');
        }

        const authConfig: AuthConfig = {
          issuer:        cfg.issuer,
          clientId:      cfg.clientId,
          redirectUri:   cfg.redirectUri  || `${window.location.origin}/api/auth/callback`,
          scope:         cfg.scope        || 'auth',
          responseType:  'code',
          loginUrl:      cfg.authUrl,
          tokenEndpoint: cfg.tokenUrl,
          oidc:          false,
          requireHttps:  false,
          showDebugInformation: false,
        };
        this.oauth.configure(authConfig);
      }),

      switchMap(() =>
        from(this.oauth.tryLogin()).pipe(
          catchError((err) => {
            console.warn('[AuthService] Token login failed:', err);
            this.configLoadFailed = true;
            return of(false);
          })
        )
      ),

      tap(() => {
        // Only redirect if login actually succeeded
        const target = sessionStorage.getItem('_redirectUri');
        if (target && this.oauth.hasValidAccessToken()) {
          sessionStorage.removeItem('_redirectUri');
          this.router.navigateByUrl(target);
        }
      }),

      map(() => undefined),

      catchError((err) => {
        console.error('[AuthService] OAuth configuration failed:', err);
        console.warn('[AuthService] App will continue in public-only mode');
        this.configLoadFailed = true;
        return of(undefined);
      })
    );
  }

  // ── queries ───────────────────────────────────────────────────────────────

  get isLoggedIn(): boolean {
    if (this.configLoadFailed) {
      return false;
    }
    return this.oauth.hasValidAccessToken();
  }

  get userInfo(): Record<string, unknown> | null {
    if (this.configLoadFailed) {
      return null;
    }
    const claims = this.oauth.getIdentityClaims();
    return claims ? claims as Record<string, unknown> : null;
  }

  get accessToken(): string | null {
    if (this.configLoadFailed) {
      return null;
    }
    return this.oauth.getAccessToken();
  }

  get isConfigured(): boolean {
    return !this.configLoadFailed;
  }

  // ── actions ───────────────────────────────────────────────────────────────

  login(): void {
    if (this.configLoadFailed) {
      console.warn('[AuthService] Cannot login: OAuth config failed to load');
      alert('Authentication is currently unavailable. Please contact support.');
      return;
    }

    sessionStorage.setItem('_redirectUri', window.location.pathname);
    this.oauth.initCodeFlow();
  }

  logout(): void {
    if (this.configLoadFailed) {
      return;
    }
    this.oauth.logOut();
  }
}
