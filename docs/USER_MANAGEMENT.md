# User Management System

## Overview

FlexFlag includes a comprehensive user management system with email/password authentication and Google OAuth integration.

## Features

### 1. User CRUD Operations

**Backend API Endpoints** (Admin only):
- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users` - List all users (with pagination)
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user details
- `DELETE /api/v1/users/:id` - Soft delete user (sets is_active to false)
- `POST /api/v1/users/:id/reset-password` - Reset user password

### 2. Password Management

**Auto-Generation**:
- Passwords can be auto-generated using secure random generation
- 12 characters with mix of letters, numbers, and special characters
- Generated passwords are shown once and must be saved securely

**Manual Password**:
- Admins can specify a custom password when creating users
- Minimum 8 characters required
- Password is hashed using bcrypt before storage

**Password Reset**:
- Admins can reset any user's password
- New password is auto-generated and displayed once
- Old password is immediately invalidated

### 3. Frontend User Interface

**User Management Page** (`/users`):
- Dashboard showing user statistics:
  - Total users
  - Active users
  - Administrators count
  - Editors count
- User table with:
  - Avatar and user details
  - Role badges (Admin, Editor, Viewer)
  - Active/Inactive status
  - Creation and last login timestamps
  - Action buttons (Reset Password, Edit, Delete)

**Create User Dialog**:
- Email input (required)
- Full name input (required)
- Password field with generate button
- Role selection (Admin, Editor, Viewer)
- Active status toggle

**Password Display**:
- Secure password display dialog
- Copy to clipboard functionality
- Warning to save password securely
- Shown after user creation or password reset

### 4. Google OAuth Integration

**OAuth Flow**:
1. User clicks "Sign in with Google" on login page
2. Redirected to Google consent screen
3. User authorizes FlexFlag to access email and profile
4. Redirected back to FlexFlag with authorization code
5. Backend exchanges code for access token
6. Backend fetches user info from Google
7. User is created or retrieved from database
8. JWT token is generated and returned
9. User is logged in automatically

**Configuration**:
- Client ID and Secret configured via environment variables
- Redirect URL configurable per environment
- State token validation for CSRF protection
- Email verification required

**Frontend Components**:
- `GoogleSignInButton` - Reusable OAuth button component
- `/auth/google/callback` - OAuth callback handler page
- Integrated into login page with divider

## User Roles

### Admin
- Full system access
- Can create, edit, and delete users
- Can reset passwords
- Can manage all flags and projects

### Editor
- Can create and modify flags
- Can create and modify projects
- Cannot manage users
- Cannot delete critical resources

### Viewer
- Read-only access
- Can view flags and projects
- Cannot make changes
- Default role for new OAuth users

## Security Features

1. **Password Hashing**: All passwords are hashed using bcrypt
2. **JWT Authentication**: Secure token-based authentication
3. **CSRF Protection**: State token validation in OAuth flow
4. **Email Verification**: OAuth users must have verified email
5. **Role-Based Access**: Endpoint protection based on user roles
6. **Soft Delete**: Users are deactivated, not permanently deleted
7. **Secure Password Generation**: Cryptographically secure random passwords

## API Usage Examples

### Create User with Auto-Generated Password

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "full_name": "John Doe",
    "role": "editor"
  }'
```

Response:
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "full_name": "John Doe",
    "role": "editor",
    "is_active": true
  },
  "password": "aB3$xY9#qR2%",
  "message": "User created successfully. Please save the password securely."
}
```

### List Users

```bash
curl -X GET "http://localhost:8080/api/v1/users?limit=50&offset=0" \
  -H "Authorization: Bearer <admin-token>"
```

### Reset User Password

```bash
curl -X POST http://localhost:8080/api/v1/users/uuid/reset-password \
  -H "Authorization: Bearer <admin-token>"
```

Response:
```json
{
  "message": "Password reset successfully",
  "password": "nP8&zQ4@tM1#"
}
```

## Frontend Usage

### User Management

```typescript
import { apiClient } from '@/lib/api';

// Load users
const users = await apiClient.getUsers();

// Create user
const response = await apiClient.createUser({
  email: 'user@example.com',
  full_name: 'John Doe',
  role: 'editor'
});

// Update user
await apiClient.updateUser(userId, {
  full_name: 'Jane Doe',
  role: 'admin',
  is_active: true
});

// Delete user
await apiClient.deleteUser(userId);

// Reset password
const { password } = await apiClient.resetUserPassword(userId);
```

### Google OAuth

```typescript
// Login page already includes Google Sign-In button
// Users just click and authenticate
// No additional frontend code needed
```

## Environment Variables

```bash
# OAuth Configuration
FLEXFLAG_AUTH_OAUTH_GOOGLE_CLIENT_ID="your-google-client-id"
FLEXFLAG_AUTH_OAUTH_GOOGLE_CLIENT_SECRET="your-google-client-secret"
FLEXFLAG_AUTH_OAUTH_GOOGLE_REDIRECT_URL="http://localhost:3000/auth/google/callback"
```

## Database Schema

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),  -- NULL for OAuth users
    full_name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'viewer',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Future Enhancements

- [ ] Multi-factor authentication (MFA)
- [ ] Password complexity requirements
- [ ] Password expiration policies
- [ ] Login attempt tracking and lockout
- [ ] Audit log for user management actions
- [ ] Email verification for manual registrations
- [ ] OAuth providers (GitHub, Microsoft, etc.)
- [ ] SSO integration (SAML, LDAP)
- [ ] Session management and revocation
- [ ] User activity tracking
