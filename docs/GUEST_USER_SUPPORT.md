# Guest User Support - Project Generation

This document explains how the GoCraft backend supports both authenticated and guest (unauthenticated) users for project generation.

## Problem Solved

Previously, the `/generate` endpoint required user authentication, preventing guest users from downloading generated projects without signing up or logging in. This created a barrier to adoption and user experience.

## Solution Overview

We've implemented multiple approaches to support guest users while maintaining the benefits of authentication for registered users:

### 1. Public Generation Endpoints

#### `/api/v1/generate` (Recommended)
- **Authentication**: Optional (uses `OptionalAuth` middleware)
- **Behavior**: 
  - If user provides valid JWT token → project is generated and could be tracked in history (if history tracking is enabled)
  - If no token or invalid token → project is generated as guest user
- **Use Case**: Modern frontend applications that want to support both guest and authenticated users seamlessly

#### `/generate` (Legacy)
- **Authentication**: None required (`GeneratePublic` handler)
- **Behavior**: Always treats user as guest, no authentication checking
- **Use Case**: Simple integrations, backward compatibility

### 2. Authentication Middleware

#### `OptionalAuth` Middleware
- Extracts user information if a valid JWT token is provided
- Continues execution as guest user if no token or invalid token
- Sets context variables:
  - `user_id`: User ID if authenticated, `nil` if guest
  - `jwt_claims`: JWT claims if authenticated, `nil` if guest  
  - `is_authenticated`: `true` if authenticated, `false` if guest

#### `RequireAuth` Middleware (Existing)
- Still used for endpoints that must have authentication
- Returns 401 error if no valid token provided

### 3. Handler Functions

#### `GenerateWithOptionalAuth` 
- Works with `OptionalAuth` middleware
- Generates projects for both authenticated and guest users
- Logs requests differently based on authentication status
- Can be extended to track history for authenticated users only

#### `GeneratePublic`
- Simple handler for guest users only
- No authentication logic at all
- Faster execution for pure guest scenarios

#### `Generate` (Existing)
- Still used for authenticated-only endpoints
- Requires valid user authentication

## API Endpoints Summary

| Endpoint | Authentication | Middleware | Handler | Use Case |
|----------|---------------|------------|---------|----------|
| `POST /api/v1/generate` | Optional | `OptionalAuth` | `GenerateWithOptionalAuth` | Modern apps (recommended) |
| `POST /generate` | None | None | `GeneratePublic` | Legacy support |
| `POST /generate/authenticated` | Required | `RequireAuth` | `Generate` | Authenticated-only features |

## Frontend Integration

### For Guest Users

```javascript
// No Authorization header needed
const response = await fetch('/api/v1/generate', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    project_name: 'my-backend',
    features: ['gin', 'postgresql', 'jwt']
  })
});

const blob = await response.blob();
// Handle download...
```

### For Authenticated Users

```javascript
// Include Authorization header if user is logged in
const headers = {
  'Content-Type': 'application/json'
};

if (userToken) {
  headers['Authorization'] = `Bearer ${userToken}`;
}

const response = await fetch('/api/v1/generate', {
  method: 'POST',
  headers,
  body: JSON.stringify({
    project_name: 'my-backend', 
    features: ['gin', 'postgresql', 'jwt']
  })
});
```

### Conditional Header Logic (Recommended Frontend Pattern)

```javascript
function generateProject(projectData) {
  const headers = {
    'Content-Type': 'application/json'
  };
  
  // Only add auth header if user is actually logged in
  const token = localStorage.getItem('authToken');
  if (token && token !== 'null' && token !== 'undefined') {
    headers['Authorization'] = `Bearer ${token}`;
  }
  
  return fetch('/api/v1/generate', {
    method: 'POST',
    headers,
    body: JSON.stringify(projectData)
  });
}
```

## Security Considerations

1. **No Sensitive Data Exposure**: Guest users cannot access any user-specific data
2. **Rate Limiting**: Applied to all endpoints regardless of authentication status
3. **Input Validation**: Same validation rules apply to both guest and authenticated requests
4. **Logging**: Different log formats help distinguish between guest and authenticated requests

## Testing

To test guest user functionality:

```bash
# Test without authentication
curl -X POST http://localhost:8080/api/v1/generate \
  -H "Content-Type: application/json" \
  -d '{"project_name":"test","features":["gin"]}' \
  --output test-project.zip

# Test with invalid token (should still work)
curl -X POST http://localhost:8080/api/v1/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer invalid-token" \
  -d '{"project_name":"test","features":["gin"]}' \
  --output test-project.zip
```

Both requests should succeed and return a downloadable zip file.