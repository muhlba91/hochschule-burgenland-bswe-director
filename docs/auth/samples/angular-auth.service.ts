import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';

/**
 * Example Angular Service for handling Flow Director Authentication.
 * 
 * NOTE: In a production environment, the 'secret' should NEVER be in the frontend.
 * The frontend should ideally call a proxy or just receive the JWT from a secure backend.
 */
@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private directorUrl = 'http://localhost:8080'; // Director Base URL
  private token: string | null = null;

  constructor(private http: HttpClient) { }

  /**
   * Performs the token exchange using the ClientID and a pre-shared API Key.
   * If the student only has the secret, they must generate the HMAC SHA256 first (see principles.md).
   */
  exchangeToken(clientID: string, apiKey: string): Observable<any> {
    return this.http.post(`${this.directorUrl}/api/v1/auth/token`, {
      clientID: clientID,
      apiKey: apiKey
    }).pipe(
      tap((res: any) => {
        this.token = res.token;
        console.log('JWT Token acquired:', this.token);
      })
    );
  }

  /**
   * Connect to the WebSocket with the acquired JWT.
   * Note: Some implementations pass the token in query params, others in headers.
   * The Director currently expects it via the specific protocol handling or header if supported by the client.
   */
  getWebSocketUrl(): string {
    return `ws://localhost:8080/api/v1/ws?token=${this.token}`;
  }
}
