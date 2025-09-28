# Troubleshooting Guide

## Issue: .env.example and .gitignore files not visible after download

### Problem Description
After downloading and extracting the generated project ZIP file, the `.env.example` and `.gitignore` files are not visible in the project directory.

### Root Cause
This is typically caused by the operating system hiding "dotfiles" (files that start with a dot) by default.

### Solutions

#### For macOS/Linux Users:

1. **Show hidden files in terminal:**
   ```bash
   ls -la
   ```
   This will show all files including hidden ones.

2. **Show hidden files in Finder (macOS):**
   - Press `Cmd + Shift + .` (period) to toggle hidden file visibility
   - Or use: `defaults write com.apple.finder AppleShowAllFiles YES && killall Finder`

3. **Show hidden files in file manager (Linux):**
   - Press `Ctrl + H` in most file managers
   - Or use `ls -la` in terminal

#### For Windows Users:

1. **Show hidden files in File Explorer:**
   - Open File Explorer
   - Click on "View" tab
   - Check "Hidden items" checkbox

2. **Show hidden files via Command Prompt:**
   ```cmd
   dir /a
   ```

### Verification Steps

1. **Use the debug endpoint:**
   Visit `http://localhost:8081/debug` to verify project generation and see all files that should be included.

2. **Manual verification:**
   ```bash
   # Extract and list all files
   unzip your-project.zip
   cd your-project
   ls -la  # On Unix/Mac
   dir /a  # On Windows
   ```

3. **Expected files:**
   ```
   .env.example     # Environment configuration template
   .gitignore       # Git ignore rules
   go.mod           # Go module definition
   cmd/             # Application entry points
   └── [project]/   # Project-specific directory
       └── main.go  # Application entry point
   internal/        # Internal packages directory
   ```

### API Verification

Use the verification endpoint to check what files are included:

```bash
curl -X POST http://localhost:8081/generate/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "projectName": "test-project",
    "features": ["postgresql", "auth"]
  }'
```

### Common Issues and Solutions

#### Issue: go.mod has invalid module name
**Solution:** The generator now creates proper module names. If you see `module go`, regenerate the project.

#### Issue: Files are in ZIP but not extracted
**Solution:** 
- Try a different extraction tool
- Check if your extraction tool preserves hidden files
- Use command line extraction: `unzip -a project.zip`

#### Issue: Browser download corrupted
**Solution:**
- Try downloading with curl:
  ```bash
  curl -X POST http://localhost:8081/generate \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer YOUR_JWT_TOKEN" \
    -d '{"projectName":"test","features":["auth"]}' \
    --output test.zip
  ```

### Debug Tools

1. **Web Debug Interface:** Visit `http://localhost:8081/debug`
2. **Verification API:** `POST /generate/verify`
3. **Manual ZIP inspection:** Use any ZIP utility to view contents before extraction

### Getting Help

If you're still experiencing issues:

1. Use the debug interface to verify file generation
2. Check the verification API response
3. Ensure your extraction tool supports hidden files
4. Try command-line extraction tools

### Example: Complete Verification Process

```bash
# 1. Generate and verify
curl -X POST http://localhost:8081/generate/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{"projectName":"test","features":["auth","postgresql"]}' | jq

# 2. Download project
curl -X POST http://localhost:8081/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{"projectName":"test","features":["auth","postgresql"]}' \
  --output test.zip

# 3. Extract and verify
unzip test.zip
cd test
ls -la

# 4. Expected output should include:
# .env.example
# .gitignore  
# go.mod
# main.go
# internal/
```

This should resolve the visibility issues with dotfiles in generated projects.