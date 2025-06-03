# Single User Configuration

The Owlistic AI application has been configured as a single-user system where the user account is managed via environment variables instead of registration functionality.

## Environment Variables

Add these variables to your `.env` file to configure the single user account:

```env
USER_USERNAME=admin
USER_EMAIL=admin@owlistic.local  
USER_PASSWORD=your_secure_password
```

### Configuration Options

- **USER_USERNAME**: The username for the single user (default: `admin`)
- **USER_EMAIL**: The email address for login (default: `admin@owlistic.local`)  
- **USER_PASSWORD**: The password for login (default: `admin123`)

⚠️ **Important**: Change the default password to something secure in production!

## How It Works

### Database Setup
- On application startup, the system automatically creates or updates the single user account
- User credentials are hashed and stored securely in the database
- If a user already exists, their credentials are updated to match the environment variables

### Authentication
- Users log in using the email and password configured in the environment variables
- The login process remains the same - only the registration functionality has been removed
- JWT tokens are still used for session management

### User Management
- No user registration endpoints are available
- The single user's details can be updated via the existing user update endpoints
- Password can be changed through the API after authentication

## Changes Made

### Backend Changes

#### Database Setup (`database/migrations.go`)
- Added `SetupSingleUser()` function that creates/updates user from environment variables
- Automatically called during database migrations
- Handles password hashing and secure storage

#### Configuration (`config/config.go`)
- Added `UserUsername`, `UserEmail`, and `UserPassword` fields
- Loads user credentials from environment variables with sensible defaults

#### Routes (`routes/users.go`)
- Removed `POST /register` endpoint
- Removed `CreateUser` function
- Registration functionality completely disabled

### Frontend Changes

#### Login Screen (`lib/screens/login_screen.dart`)
- Removed "Don't have an account? Register now" link
- Cleaner login interface with only email, password, and remember me

#### View Models (`lib/viewmodel/login_viewmodel.dart`)
- Removed `navigateToRegister` method
- Simplified interface without registration navigation

#### Router (`lib/core/router.dart`)
- Removed `/register` route
- Removed register screen import
- Simplified redirect logic without registration paths

#### Provider (`lib/providers/login_provider.dart`)
- Removed `navigateToRegister` implementation
- Cleaner authentication flow

## Security Considerations

### Password Security
- Passwords are hashed using bcrypt with default cost
- Password hashes are never exposed in API responses
- Environment variables should be kept secure

### Access Control
- Single user has full access to all application features
- All existing authorization checks remain in place
- JWT tokens expire based on configured duration

### Environment Protection
- Keep `.env` file secure and never commit it to version control
- Use strong passwords in production environments
- Consider using secrets management for production deployments

## Migration from Multi-User

If migrating from a multi-user setup:

1. **Backup existing users**: Export user data if needed for reference
2. **Set environment variables**: Configure the single user credentials
3. **Restart application**: The system will automatically update the user account
4. **Verify login**: Test login with the configured credentials

### Existing Data
- All existing notes, tasks, notebooks, and other data remain intact
- The system will use the first existing user and update their credentials
- If no users exist, a new user will be created

## API Changes

### Removed Endpoints
- `POST /api/v1/register` - User registration (removed)

### Unchanged Endpoints
- `POST /api/v1/login` - User login (works with configured credentials)
- `GET /api/v1/users/:id` - Get user details (works for the single user)
- `PUT /api/v1/users/:id` - Update user details (works for the single user)
- `PUT /api/v1/users/:id/password` - Change password (works for the single user)

## Development vs Production

### Development Setup
```env
USER_USERNAME=dev
USER_EMAIL=dev@localhost
USER_PASSWORD=dev123
```

### Production Setup
```env
USER_USERNAME=admin
USER_EMAIL=admin@yourdomain.com
USER_PASSWORD=very_secure_random_password_here
```

## Troubleshooting

### Cannot Login
1. Check that environment variables are set correctly
2. Verify the application restarted after changing environment variables
3. Check logs for user creation/update messages
4. Ensure the email format is valid

### Password Not Working
1. Environment variables are case-sensitive
2. Make sure there are no extra spaces in the password
3. Check if the password was updated (restart required)
4. Verify the password meets any validation requirements

### Database Issues
1. Check database migrations completed successfully
2. Verify the users table exists and has the correct schema
3. Look for error messages during startup
4. Ensure database connection is working

## Examples

### Docker Compose
```yaml
services:
  backend:
    environment:
      - USER_USERNAME=admin
      - USER_EMAIL=admin@mycompany.com
      - USER_PASSWORD=${ADMIN_PASSWORD}
```

### Kubernetes Secret
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: user-credentials
data:
  USER_USERNAME: YWRtaW4=  # base64 encoded 'admin'
  USER_EMAIL: YWRtaW5AbXljb21wYW55LmNvbQ==  # base64 encoded email
  USER_PASSWORD: c2VjdXJlX3Bhc3N3b3Jk  # base64 encoded password
```

The single-user configuration provides a simpler, more secure setup for personal or small team deployments while maintaining all the functionality of the full application.