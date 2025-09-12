# Framework Field Examples

## Problem Fixed
The frontend was sending framework as a separate field, but the backend expected it in the features array.

## Solution
The backend now supports both request formats:

### Format 1: Framework in Features Array (Traditional)
```json
{
  "projectName": "my-api",
  "features": ["gin", "postgresql", "auth"]
}
```

### Format 2: Framework as Separate Field (New)
```json
{
  "projectName": "my-api",
  "framework": "fiber",
  "features": ["mongodb", "auth"]
}
```

## Backend Processing
The backend automatically merges the framework field into the features array:

```javascript
// Frontend sends:
{
  "projectName": "test",
  "framework": "fiber", 
  "features": ["mongodb"]
}

// Backend processes as:
{
  "projectName": "test",
  "features": ["fiber", "mongodb"]  // Framework merged into features
}
```

## Validation Examples

### Valid Requests
```json
// Fiber with MongoDB
{
  "projectName": "my-app",
  "framework": "fiber",
  "features": ["mongodb", "redis", "auth"]
}
// Result: ✅ Valid - generates Fiber project with MongoDB

// Echo with PostgreSQL  
{
  "projectName": "my-app",
  "framework": "echo", 
  "features": ["postgresql", "gorm", "auth"]
}
// Result: ✅ Valid - generates Echo project with PostgreSQL

// Gin (traditional format)
{
  "projectName": "my-app",
  "features": ["gin", "mysql", "auth"]
}
// Result: ✅ Valid - generates Gin project with MySQL
```

### Invalid Requests
```json
// Multiple frameworks
{
  "projectName": "my-app",
  "framework": "fiber",
  "features": ["gin", "postgresql"]  // Conflict: fiber + gin
}
// Result: ❌ Error - Multiple web frameworks selected

// Multiple primary databases
{
  "projectName": "my-app", 
  "framework": "echo",
  "features": ["postgresql", "mongodb"]  // Conflict: postgresql + mongodb
}
// Result: ❌ Error - Multiple primary databases selected
```

## Frontend Implementation
The frontend can now send either format:

```javascript
// Option 1: Separate framework field (recommended)
const request = {
  projectName: projectName,
  framework: selectedFramework,  // "gin", "echo", or "fiber"
  features: selectedFeatures     // ["postgresql", "auth", "redis"]
};

// Option 2: Framework in features (traditional)
const request = {
  projectName: projectName,
  features: [selectedFramework, ...selectedFeatures]  // ["gin", "postgresql", "auth"]
};
```

## Generated Projects
Now correctly generates the selected framework:

- **Fiber selected** → Generates Fiber imports and initialization
- **Echo selected** → Generates Echo imports and initialization  
- **Gin selected** → Generates Gin imports and initialization

The bug where Fiber/Echo selection generated Gin projects is now fixed! 🎉