import { Injectable }                        from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';
import { Observable }                       from 'rxjs';
import { map }                             from 'rxjs/operators';

/**
 * Low-level, reusable API client.
 *
 * Design goals
 * ------------
 * 1. Every public method is generic so the caller can specify the response shape.
 * 2. A base URL is **not** hard-coded; callers pass full paths (e.g. `/api/...`).
 *    The Angular dev-server proxy (proxy.conf.json) rewrites `/api/*` → backend.
 * 3. To add a new "resource" (e.g. users, products) you can either:
 *      a) Inject ApiService directly and call get/post/put/delete, or
 *      b) Create a domain service that injects ApiService and wraps it.
 *
 * Example
 * -------
 *   constructor(private api: ApiService) {}
 *
 *   getUsers(): Observable<User[]> {
 *     return this.api.get<User[]>('/api/users');
 *   }
 */
@Injectable({ providedIn: 'root' })
export class ApiService {

  constructor(private http: HttpClient) {}

  // ── Generic CRUD ──────────────────────────────────────────────────────────

  /** HTTP GET – returns the parsed JSON body. */
  get<T>(url: string, params?: HttpParams): Observable<T> {
    return this.http.get<T>(url, { params });
  }

  /** HTTP POST – sends `body` as JSON. */
  post<T>(url: string, body: unknown, headers?: HttpHeaders): Observable<T> {
    return this.http.post<T>(url, body, { headers });
  }

  /** HTTP PUT – replaces the resource at `url`. */
  put<T>(url: string, body: unknown, headers?: HttpHeaders): Observable<T> {
    return this.http.put<T>(url, body, { headers });
  }

  /** HTTP PATCH – partial update. */
  patch<T>(url: string, body: unknown, headers?: HttpHeaders): Observable<T> {
    return this.http.patch<T>(url, body, { headers });
  }

  /** HTTP DELETE – returns the parsed response body (if any). */
  delete<T>(url: string, params?: HttpParams): Observable<T> {
    return this.http.delete<T>(url, { params });
  }

  // ── Convenience: flat key→value config loader (used by OAuth bootstrap) ───

  /**
   * Fetches a flat JSON object from `url` and returns it as a plain record.
   * This is the pattern expected by `/api/oauth/config`.
   */
  getConfig(url: string): Observable<Record<string, string>> {
    return this.http.get<Record<string, string>>(url);
  }
}
