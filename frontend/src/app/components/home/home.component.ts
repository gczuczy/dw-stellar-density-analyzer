import { Component } from '@angular/core';

@Component({
  selector:   'app-home',
  standalone: true,
  template: `
    <div class="row justify-content-center mt-4">
      <div class="col-md-8">
        <div class="card p-4">
          <h1 class="card-title">Welcome</h1>
          <p class="card-text">
            This is the public home page of the Stellar Density Analyzer.
            No authentication is required to view this page.
          </p>
          <p class="card-text text-secondary">
            Use the navigation bar above to explore the different sections.
            Pages marked as <em>login-protected</em> will redirect you through
            the OAuth2 PKCE authentication flow before granting access.
          </p>
        </div>
      </div>
    </div>
  `
})
export class HomeComponent {}
