import { Injectable }              from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable }              from 'rxjs';

/**
 * Generic HTTP client service with typed methods.
 * Extend this or use directly in your components.
 */
@Injectable({ providedIn: 'root' })
export class ApiService {

  constructor(private http: HttpClient) {}

  /**
   * Fetches configuration from a backend endpoint.
   * Returns the JSON response as a Record<string, string>.
   */
  getConfig(url: string): Observable<Record<string, string>> {
    return this.http.get<Record<string, string>>(url);
  }

  // ── Generic typed HTTP methods ────────────────────────────────────────────

  get<T>(url: string, options?: { headers?: HttpHeaders }): Observable<T> {
    return this.http.get<T>(url, options);
  }

  post<T>(url: string, body: unknown, options?: { headers?: HttpHeaders }): Observable<T> {
    return this.http.post<T>(url, body, options);
  }

  put<T>(url: string, body: unknown, options?: { headers?: HttpHeaders }): Observable<T> {
    return this.http.put<T>(url, body, options);
  }

  delete<T>(url: string, options?: { headers?: HttpHeaders }): Observable<T> {
    return this.http.delete<T>(url, options);
  }

  patch<T>(url: string, body: unknown, options?: { headers?: HttpHeaders }): Observable<T> {
    return this.http.patch<T>(url, body, options);
  }
}
