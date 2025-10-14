# Security Policy

## Supported Versions

We actively support the following versions of GoCraft with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of GoCraft seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### How to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to: **[INSERT SECURITY EMAIL]**

You should receive a response within 48 hours. If for some reason you do not, please follow up via email to ensure we received your original message.

### What to Include

Please include the following information in your report:

- **Type of issue** (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- **Full paths of source file(s)** related to the manifestation of the issue
- **The location of the affected source code** (tag/branch/commit or direct URL)
- **Any special configuration required** to reproduce the issue
- **Step-by-step instructions to reproduce the issue**
- **Proof-of-concept or exploit code** (if possible)
- **Impact of the issue**, including how an attacker might exploit the issue

This information will help us triage your report more quickly.

### Preferred Languages

We prefer all communications to be in English.

## Security Response Process

1. **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours.

2. **Initial Assessment**: We will perform an initial assessment of the reported vulnerability within 5 business days.

3. **Investigation**: Our security team will investigate the issue and determine:
   - Whether the issue is a valid security vulnerability
   - The severity and impact of the vulnerability
   - Which versions are affected
   - What fixes are required

4. **Resolution Timeline**: 
   - **Critical vulnerabilities**: Patches within 7 days
   - **High severity**: Patches within 30 days
   - **Medium/Low severity**: Patches in next regular release cycle

5. **Disclosure**: We follow responsible disclosure practices:
   - We will work with you to understand the issue fully
   - We will keep you informed of our progress
   - We will coordinate the disclosure timeline with you
   - We will credit you in the security advisory (unless you prefer to remain anonymous)

## Security Best Practices

### For Users

When using GoCraft, please follow these security best practices:

1. **Keep Updated**: Always use the latest supported version
2. **Environment Variables**: Never commit sensitive data like API keys or passwords to version control
3. **Generated Code Review**: Review generated code before deploying to production
4. **Access Control**: Implement proper authentication and authorization in generated applications
5. **Input Validation**: Ensure proper input validation in your applications
6. **HTTPS**: Always use HTTPS in production environments

### For Contributors

When contributing to GoCraft:

1. **Dependency Security**: Keep dependencies updated and scan for vulnerabilities
2. **Code Review**: All code changes require review before merging
3. **Static Analysis**: Use static analysis tools to identify potential security issues
4. **Input Sanitization**: Properly sanitize and validate all user inputs
5. **Error Handling**: Avoid exposing sensitive information in error messages
6. **Secrets Management**: Never hardcode secrets or credentials

## Security Features

GoCraft includes several security features:

### Template Security
- Input validation for all template parameters
- Sanitization of user-provided values
- Protection against template injection attacks

### Generated Code Security
- Secure defaults for authentication and authorization
- HTTPS enforcement options
- Input validation middleware
- SQL injection prevention (when using supported ORMs)
- CORS configuration options

### Build Security
- Dependency vulnerability scanning
- Static code analysis integration
- Secure build pipeline practices

## Known Security Considerations

### Template Processing
- Templates are processed server-side and should be treated as trusted code
- User input is validated before template processing
- Generated code should be reviewed before production deployment

### File System Access
- GoCraft requires file system write access to generate projects
- Generated files should be reviewed before execution
- Temporary files are cleaned up after generation

### Network Access
- GoCraft may download dependencies during code generation
- Ensure network security policies allow necessary connections
- Consider using private package repositories in restricted environments

## Security Updates

Security updates will be:

1. **Announced** via GitHub Security Advisories
2. **Tagged** with clear version numbers following semantic versioning
3. **Documented** with detailed changelog entries
4. **Communicated** through our official channels

## Vulnerability Disclosure Policy

We believe in responsible disclosure and will:

- Acknowledge security researchers who report vulnerabilities
- Provide credit in security advisories (unless anonymity is requested)
- Work collaboratively to understand and resolve issues
- Maintain confidentiality until patches are available
- Coordinate disclosure timing to protect users

## Contact Information

For security-related questions or concerns:

- **Security Email**: [INSERT SECURITY EMAIL]
- **General Contact**: [INSERT GENERAL EMAIL]
- **GitHub Issues**: For non-security related bugs and features only

## Legal

This security policy is subject to our [Terms of Service](https://github.com/telman03/gocraft-backend) and [Privacy Policy](https://github.com/telman03/gocraft-backend).

By reporting a vulnerability, you agree to:
- Not publicly disclose the issue until we have had a chance to address it
- Not exploit the vulnerability for malicious purposes
- Act in good faith to avoid privacy violations and disruption to others

We commit to:
- Respond to your report in a timely manner
- Keep you informed of our progress
- Credit your contribution (unless you prefer anonymity)
- Not pursue legal action against researchers who follow this policy

---

Thank you for helping keep GoCraft and our users safe!