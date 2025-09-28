# Repository Settings Documentation

This document outlines the recommended GitHub repository settings for the GoCraft project to ensure proper open-source governance and community management.

## General Settings

### Repository Details
- **Description**: "A powerful Go backend generator that creates production-ready scaffolds with configurable features"
- **Website**: Link to documentation site (if available)
- **Topics**: `go`, `backend`, `generator`, `scaffold`, `web-framework`, `gin`, `fiber`, `echo`, `postgresql`, `mysql`, `sqlite`
- **Include in the home page**: ✅ Checked

### Features
- **Wikis**: ❌ Disabled (use docs/ directory instead)
- **Issues**: ✅ Enabled
- **Sponsorships**: ✅ Enabled (if applicable)
- **Preserve this repository**: ✅ Enabled
- **Discussions**: ✅ Enabled (recommended for community Q&A)

### Pull Requests
- **Allow merge commits**: ✅ Enabled
- **Allow squash merging**: ✅ Enabled (default)
- **Allow rebase merging**: ✅ Enabled
- **Always suggest updating pull request branches**: ✅ Enabled
- **Allow auto-merge**: ✅ Enabled
- **Automatically delete head branches**: ✅ Enabled

## Branch Protection Rules

### Main Branch Protection
Configure the following rules for the `main` branch:

#### Protect matching branches
- **Require a pull request before merging**: ✅ Enabled
  - **Require approvals**: 1 (minimum)
  - **Dismiss stale reviews**: ✅ Enabled
  - **Require review from code owners**: ✅ Enabled (if CODEOWNERS file exists)
  - **Restrict pushes that create files**: ❌ Disabled
  - **Allow specified actors to bypass required pull requests**: Configure for maintainers only

#### Require status checks to pass
- **Require branches to be up to date**: ✅ Enabled
- **Status checks**: 
  - `test (1.20)` - Go 1.20 tests
  - `test (1.21)` - Go 1.21 tests
  - `build` - Build verification
  - `lint` - Code linting
  - `security` - Security scanning

#### Additional Restrictions
- **Restrict pushes**: Configure for maintainers and admins only
- **Allow force pushes**: ❌ Disabled
- **Allow deletions**: ❌ Disabled

## Security Settings

### Security & Analysis
- **Dependency graph**: ✅ Enabled
- **Dependabot alerts**: ✅ Enabled
- **Dependabot security updates**: ✅ Enabled
- **Code scanning**: ✅ Enabled (CodeQL)
- **Secret scanning**: ✅ Enabled

### Dependabot Configuration
Create `.github/dependabot.yml`:
```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10
```

## Issue and PR Templates

### Issue Templates
- **Bug Report**: Comprehensive bug reporting template
- **Feature Request**: Structured feature request template  
- **Documentation**: Documentation improvement template
- **Config**: Disable blank issues, provide helpful links

### Pull Request Template
- Comprehensive checklist covering code quality, testing, documentation
- Specific sections for template and generation testing
- Security and breaking change considerations

## GitHub Actions Workflows

### CI Workflow (`ci.yml`)
- **Triggers**: Push to main/develop, PRs to main/develop
- **Jobs**: 
  - Test (Go 1.20, 1.21)
  - Build verification
  - Linting (golangci-lint)
  - Security scanning (gosec)
- **Coverage**: Upload to Codecov

### Release Workflow (`release.yml`)
- **Triggers**: Version tags (v*)
- **Builds**: Multi-platform binaries (Linux, macOS, Windows)
- **Artifacts**: Automated release creation with binaries

## Labels Configuration

### Standard Labels
Create the following labels for better issue organization:

#### Type Labels
- `bug` (🐛) - Something isn't working
- `enhancement` (✨) - New feature or request
- `documentation` (📚) - Improvements or additions to documentation
- `question` (❓) - Further information is requested
- `help wanted` (🙋) - Extra attention is needed
- `good first issue` (👋) - Good for newcomers

#### Priority Labels
- `priority: critical` (🔥) - Critical priority
- `priority: high` (⬆️) - High priority
- `priority: medium` (➡️) - Medium priority
- `priority: low` (⬇️) - Low priority

#### Framework Labels
- `framework: gin` - Related to Gin framework
- `framework: fiber` - Related to Fiber framework
- `framework: echo` - Related to Echo framework

#### Component Labels
- `component: templates` - Template-related issues
- `component: database` - Database-related issues
- `component: auth` - Authentication-related issues
- `component: api` - API-related issues

## Community Health Files

Ensure the following files are present in the repository root:
- `LICENSE` - MIT License
- `CONTRIBUTING.md` - Contribution guidelines
- `CODE_OF_CONDUCT.md` - Community standards
- `SECURITY.md` - Security policy

## Recommended Integrations

### Code Quality
- **Codecov**: Code coverage reporting
- **CodeQL**: Security analysis
- **Dependabot**: Dependency updates

### Community
- **All Contributors**: Recognize contributors
- **Stale Bot**: Manage stale issues and PRs

## Maintenance

### Regular Tasks
- Review and update branch protection rules quarterly
- Update CI/CD workflows as Go versions change
- Review and update issue/PR templates based on community feedback
- Monitor security alerts and update dependencies

### Metrics to Monitor
- Issue response time
- PR review time
- Test coverage percentage
- Security vulnerability count
- Community engagement (stars, forks, discussions)

## Notes

- Replace `your-org/gocraft` with actual repository path in config files
- Adjust reviewer requirements based on team size
- Consider enabling GitHub Discussions for community Q&A
- Review security settings regularly and enable new features as they become available