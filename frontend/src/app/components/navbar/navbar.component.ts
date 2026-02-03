import { Component, OnInit }              from '@angular/core';
import { RouterLink, RouterLinkActive }   from '@angular/router';
import { NgbDropdownModule }              from '@ng-bootstrap/ng-bootstrap';
import { AuthService }                    from '../../auth/auth.service';

@Component({
  selector:   'app-navbar',
  standalone: true,
  imports: [
    RouterLink,
    RouterLinkActive,
    NgbDropdownModule,
  ],
  template: `
  <nav class="navbar navbar-expand-lg navbar-bg shadow-sm fixed-top">
    <div class="container-fluid">

      <!-- ── Brand ───────────────────────────────────────────────────── -->
      <a class="navbar-brand" routerLink="/">Stellar Density Analyzer</a>

      <!-- ── Toggler (mobile) ────────────────────────────────────────── -->
      <button class="navbar-toggler" type="button"
              data-bs-toggle="collapse" data-bs-target="#navbarNav"
              aria-controls="navbarNav" aria-expanded="false"
              aria-label="Toggle navigation">
        <span class="navbar-toggler-icon"></span>
      </button>

      <!-- ── Nav links ───────────────────────────────────────────────── -->
      <div class="collapse navbar-collapse" id="navbarNav">
        <ul class="navbar-nav">

          <!-- 3rd menu – barfoo – public ─────────────────────────────── -->
          <li class="nav-item">
            <a class="nav-link" routerLink="/barfoo" routerLinkActive="active">Barfoo</a>
          </li>

          <!-- 2nd menu – foobar – login-protected (guard on the route) ─ -->
          <li class="nav-item">
            <a class="nav-link" routerLink="/foobar" routerLinkActive="active">Foobar</a>
          </li>

          <!-- 4th menu – Side Menu dropdown – login-protected ─────────── -->
          <li class="nav-item" ngbDropdown>
            <a class="nav-link dropdown-toggle" ngbDropdownToggle
               [class.active]="isSidemenuActive">
              Side Menu
            </a>
            <div ngbDropdownMenu>
              <a ngbDropdownItem routerLink="/sidemenu/alpha">Alpha</a>
              <a ngbDropdownItem routerLink="/sidemenu/beta" >Beta</a>
            </div>
          </li>
        </ul>

        <!-- ── Right-aligned: Settings dropdown / Login button ─────── -->
        <ul class="navbar-nav ms-auto">
          @if (authService.isLoggedIn) {
            <li class="nav-item" ngbDropdown placement="bottom-end">
              <button class="btn btn-outline-secondary btn-sm" ngbDropdownToggle>
                ⚙ Settings
              </button>
              <div ngbDropdownMenu>
                <a ngbDropdownItem routerLink="/settings">Settings</a>
                <div class="dropdown-divider"></div>
                <button ngbDropdownItem (click)="logout()">Logout</button>
              </div>
            </li>
          } @else {
            <li class="nav-item">
              <button class="btn btn-primary btn-sm" (click)="login()">Login</button>
            </li>
          }
        </ul>
      </div>
    </div>
  </nav>
  `,
  styles: [`
    :host { display: block; }

    .navbar-bg {
      background-color: var(--bs-navbar-bg, var(--bs-body-bg));
      color: var(--bs-body-color);
    }

    .nav-link {
      color: var(--bs-body-color);
      &:hover  { color: var(--bs-primary); }
      &.active { font-weight: 600; color: var(--bs-primary) !important; }
    }
  `]
})
export class NavbarComponent implements OnInit {
  isSidemenuActive = false;

  constructor(public authService: AuthService) {}

  ngOnInit(): void {
    const check = () => {
      this.isSidemenuActive = window.location.pathname.startsWith('/sidemenu');
    };
    check();
    window.addEventListener('popstate', check);
  }

  login():  void { this.authService.login(); }
  logout(): void { this.authService.logout(); }
}
