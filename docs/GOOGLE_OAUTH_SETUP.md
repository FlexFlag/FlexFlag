# Google OAuth Setup Guide

This guide explains how to set up Google OAuth authentication for FlexFlag.

## Prerequisites

- Google Cloud Platform account
- FlexFlag backend running
- FlexFlag frontend running

## Step 1: Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Navigate to "APIs & Services" > "Credentials"

## Step 2: Configure OAuth Consent Screen

1. Click on "OAuth consent screen" in the left sidebar
2. Select "External" user type (or "Internal" if using Google Workspace)
3. Fill in the required information:
   - App name: FlexFlag
   - User support email: your email
   - Developer contact email: your email
4. Add scopes:
   - `.../auth/userinfo.email`
   - `.../auth/userinfo.profile`
5. Save and continue

## Step 3: Create OAuth 2.0 Credentials

1. Go to "Credentials" tab
2. Click "Create Credentials" > "OAuth client ID"
3. Select "Web application" as the application type
4. Configure:
   - Name: FlexFlag OAuth Client
   - Authorized JavaScript origins:
     - `http://localhost:3000` (for development)
     - Your production frontend URL
   - Authorized redirect URIs:
     - `http://localhost:3000/auth/google/callback` (for development)
     - `https://your-domain.com/auth/google/callback` (for production)
5. Click "Create"
6. Save your Client ID and Client Secret

## Step 4: Configure FlexFlag Backend

### Option 1: Environment Variables

Set the following environment variables:

```bash
export FLEXFLAG_AUTH_OAUTH_GOOGLE_CLIENT_ID="your-client-id"
export FLEXFLAG_AUTH_OAUTH_GOOGLE_CLIENT_SECRET="your-client-secret"
export FLEXFLAG_AUTH_OAUTH_GOOGLE_REDIRECT_URL="http://localhost:3000/auth/google/callback"
```

### Option 2: Configuration File

Create or update `config.yaml`:

```yaml
auth:
  jwt_secret: "your-secret-key"
  oauth:
    google:
      client_id: "your-client-id"
      client_secret: "your-client-secret"
      redirect_url: "http://localhost:3000/auth/google/callback"
```

## Step 5: Start the Application

1. Start the backend server:
   ```bash
   make run
   ```

2. Start the frontend:
   ```bash
   cd ui && npm run dev
   ```

## Step 6: Test Google Sign-In

1. Navigate to `http://localhost:3000/login`
2. Click "Sign in with Google"
3. You will be redirected to Google's consent screen
4. Grant permissions
5. You will be redirected back to FlexFlag and automatically logged in

## OAuth Flow

The OAuth authentication flow works as follows:

1. User clicks "Sign in with Google" button
2. Frontend redirects to `/api/v1/auth/google/login`
3. Backend generates a state token and redirects to Google OAuth
4. User authenticates with Google and grants permissions
5. Google redirects to `/auth/google/callback` with authorization code
6. Backend exchanges code for access token
7. Backend fetches user info from Google
8. Backend creates or retrieves user from database
9. Backend generates JWT token and returns to frontend
10. Frontend stores token and redirects to dashboard

## Security Considerations

1. **State Parameter**: Used for CSRF protection. The backend generates and validates state tokens.

2. **HTTPS**: In production, always use HTTPS for redirect URIs.

3. **Client Secret**: Never expose your client secret in frontend code. It should only be in backend configuration.

4. **Token Storage**: JWT tokens are stored in localStorage. Consider using httpOnly cookies for production.

5. **Scope Minimal**: Only request the scopes you need (email and profile).

## Troubleshooting

### "redirect_uri_mismatch" Error

- Ensure the redirect URI in your Google Console matches exactly what's configured in FlexFlag
- Check for trailing slashes and http vs https

### "invalid_client" Error

- Verify your Client ID and Client Secret are correct
- Check that OAuth credentials are enabled in Google Console

### Users Not Being Created

- Check backend logs for errors
- Verify database connection
- Ensure user email is verified in Google account

### Token Not Persisting

- Check browser localStorage
- Verify CORS settings allow credentials
- Check that JWT secret is configured

## Production Deployment

For production deployment:

1. Update redirect URIs to use your production domain
2. Use environment variables for sensitive data
3. Enable HTTPS
4. Consider using Redis for state storage instead of in-memory map
5. Implement token refresh mechanism
6. Add rate limiting to OAuth endpoints
7. Monitor OAuth flow for suspicious activity

## API Endpoints

- `GET /api/v1/auth/google/login` - Initiate OAuth flow
- `GET /api/v1/auth/google/callback` - Handle OAuth callback

## Additional Resources

- [Google OAuth 2.0 Documentation](https://developers.google.com/identity/protocols/oauth2)
- [OAuth 2.0 Best Practices](https://tools.ietf.org/html/draft-ietf-oauth-security-topics)
