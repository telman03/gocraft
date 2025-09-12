# Frontend Integration Guide - Template Validation System

## 🎯 Overview for Frontend Team

The backend now has a **Template Validation System** that prevents users from selecting conflicting features (like MySQL + PostgreSQL together). The frontend should integrate with this system to provide **real-time validation feedback** and **prevent invalid submissions**.

## 🚨 Key Validation Rules to Handle in UI

### 1. **Database Conflicts** (Most Important)
```javascript
// ❌ INVALID - Multiple relational databases
{
  "features": ["mysql", "postgresql", "sqlite"]
  // Error: "You can't have multiple relational databases in one project"
}

// ✅ VALID - One relational + others
{
  "features": ["postgresql", "mongodb", "redis"]
  // OK: Different database types can coexist
}
```

### 2. **Framework Conflicts**
```javascript
// ❌ INVALID - Multiple web frameworks
{
  "features": ["gin", "echo", "fiber"]
  // Error: "Choose only one web framework for your project"
}

// ✅ VALID - Single framework selection
{
  "features": ["fiber", "postgresql", "auth"]
  // OK: Fiber is now fully supported
}
```

### 3. **ORM Conflicts**
```javascript
// ❌ INVALID - Multiple ORMs
{
  "features": ["gorm", "sqlc"]
  // Error: "You can only use one ORM (GORM or SQLC), not both"
}
```

## 🔌 API Endpoints for Frontend

### 1. Get Supported Features (No Auth Required)
```http
GET /features
```

**Use this to:**
- Build your feature selection UI
- Show feature categories and descriptions
- Display conflict rules to users

**Response Structure:**
```json
{
  "categories": {
    "databases": ["mysql", "postgresql", "sqlite", "mongodb", "redis"],
    "frameworks": ["gin", "echo", "fiber"],
    "orms": ["gorm", "sqlc"],
    "auth": ["auth", "oauth2"],
    "ai": ["openai", "claude", "openrouter"],
    "communication": ["grpc", "websocket"],
    "devops": ["dockerfile", "docker-compose", "makefile"],
    "documentation": ["swagger", "postman", "readme"]
  },
  "descriptions": {
    "mysql": "MySQL relational database",
    "postgresql": "PostgreSQL relational database",
    "gin": "Gin web framework - fast and minimalist",
    "echo": "Echo web framework - high performance",
    "fiber": "Fiber web framework - Express-inspired, built on Fasthttp",
    // ... more descriptions
  },
  "conflict_rules": {
    "databases": {
      "rule": "Only one relational database allowed",
      "allowed": "MySQL OR PostgreSQL OR SQLite (+ optional MongoDB/Redis)",
      "forbidden": "Multiple relational databases together"
    },
    // ... more rules
  }
}
```

### 2. Validate Features (Requires Auth)
```http
POST /generate/validate
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "projectName": "my-api",
  "features": ["mysql", "postgresql", "gin"]
}
```

**Use this for:**
- Real-time validation as user selects features
- Pre-validation before allowing generation
- Showing what dependencies will be auto-added

**Success Response (HTTP 200):**
```json
{
  "project_name": "my-api",
  "original_features": ["gin", "postgresql", "auth"],
  "validation_result": {
    "is_valid": true,
    "adjusted_features": ["gin", "postgresql", "auth", "env", "gitignore", "main", "middleware"],
    "warnings": [
      "Authentication selected. Adding middleware for security."
    ],
    "added_dependencies": ["env", "gitignore", "main", "middleware"]
  }
}
```

**Error Response (HTTP 400):**
```json
{
  "project_name": "my-api",
  "original_features": ["mysql", "postgresql", "gin"],
  "validation_result": {
    "is_valid": false,
    "errors": [
      {
        "message": "Multiple relational databases selected: mysql, postgresql",
        "conflicts": ["mysql", "postgresql"],
        "suggestions": [
          "Choose only one relational database (MySQL, PostgreSQL, or SQLite)",
          "MongoDB and Redis can be used alongside a relational database"
        ]
      }
    ]
  }
}
```

## 🎨 Frontend Implementation Suggestions

### 1. **Feature Selection UI with Real-time Validation**

```javascript
// Example React/Vue component logic
const FeatureSelector = () => {
  const [selectedFeatures, setSelectedFeatures] = useState([]);
  const [validationResult, setValidationResult] = useState(null);
  const [isValidating, setIsValidating] = useState(false);

  // Validate whenever features change
  useEffect(() => {
    if (selectedFeatures.length > 0) {
      validateFeatures();
    }
  }, [selectedFeatures]);

  const validateFeatures = async () => {
    setIsValidating(true);
    try {
      const response = await fetch('/generate/validate', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          projectName: projectName,
          features: selectedFeatures
        })
      });
      
      const result = await response.json();
      setValidationResult(result.validation_result);
    } catch (error) {
      console.error('Validation failed:', error);
    }
    setIsValidating(false);
  };

  const handleFeatureToggle = (feature) => {
    const newFeatures = selectedFeatures.includes(feature)
      ? selectedFeatures.filter(f => f !== feature)
      : [...selectedFeatures, feature];
    
    setSelectedFeatures(newFeatures);
  };

  return (
    <div>
      {/* Feature Selection */}
      <FeatureCategories 
        onFeatureToggle={handleFeatureToggle}
        selectedFeatures={selectedFeatures}
        validationResult={validationResult}
      />
      
      {/* Validation Feedback */}
      <ValidationFeedback 
        result={validationResult}
        isValidating={isValidating}
      />
      
      {/* Generate Button */}
      <GenerateButton 
        disabled={!validationResult?.is_valid || isValidating}
        features={validationResult?.adjusted_features || selectedFeatures}
      />
    </div>
  );
};
```

### 2. **Validation Feedback Component**

```javascript
const ValidationFeedback = ({ result, isValidating }) => {
  if (isValidating) {
    return <div className="validation-loading">Validating features...</div>;
  }

  if (!result) return null;

  return (
    <div className="validation-feedback">
      {/* Errors */}
      {result.errors?.map((error, index) => (
        <div key={index} className="validation-error">
          <div className="error-icon">❌</div>
          <div>
            <div className="error-message">{error.message}</div>
            <div className="error-suggestions">
              {error.suggestions.map((suggestion, i) => (
                <div key={i} className="suggestion">💡 {suggestion}</div>
              ))}
            </div>
          </div>
        </div>
      ))}

      {/* Warnings */}
      {result.warnings?.map((warning, index) => (
        <div key={index} className="validation-warning">
          <div className="warning-icon">⚠️</div>
          <div>{warning}</div>
        </div>
      ))}

      {/* Auto-added Dependencies */}
      {result.added_dependencies?.length > 0 && (
        <div className="validation-info">
          <div className="info-icon">📦</div>
          <div>
            <strong>Auto-added:</strong> {result.added_dependencies.join(', ')}
          </div>
        </div>
      )}

      {/* Success State */}
      {result.is_valid && (
        <div className="validation-success">
          <div className="success-icon">✅</div>
          <div>Configuration is valid! Ready to generate.</div>
        </div>
      )}
    </div>
  );
};
```

### 3. **Smart Feature Categories with Conflict Prevention**

```javascript
const DatabaseCategory = ({ selectedFeatures, onFeatureToggle, validationResult }) => {
  const relationalDBs = ['mysql', 'postgresql', 'sqlite'];
  const selectedRelational = selectedFeatures.filter(f => relationalDBs.includes(f));
  
  return (
    <div className="feature-category">
      <h3>Databases</h3>
      
      {/* Relational Databases */}
      <div className="feature-group">
        <h4>Relational (Choose One)</h4>
        {relationalDBs.map(db => (
          <FeatureOption
            key={db}
            feature={db}
            selected={selectedFeatures.includes(db)}
            disabled={selectedRelational.length > 0 && !selectedFeatures.includes(db)}
            onToggle={onFeatureToggle}
            conflicted={hasConflict(db, validationResult)}
          />
        ))}
      </div>
      
      {/* NoSQL Databases */}
      <div className="feature-group">
        <h4>NoSQL & Cache (Optional)</h4>
        {['mongodb', 'redis'].map(db => (
          <FeatureOption
            key={db}
            feature={db}
            selected={selectedFeatures.includes(db)}
            onToggle={onFeatureToggle}
            conflicted={hasConflict(db, validationResult)}
          />
        ))}
      </div>
    </div>
  );
};

const FeatureOption = ({ feature, selected, disabled, conflicted, onToggle }) => (
  <label className={`feature-option ${conflicted ? 'conflicted' : ''} ${disabled ? 'disabled' : ''}`}>
    <input
      type="checkbox"
      checked={selected}
      disabled={disabled}
      onChange={() => onToggle(feature)}
    />
    <span className="feature-name">{feature}</span>
    {conflicted && <span className="conflict-indicator">⚠️</span>}
  </label>
);
```

## 🎯 User Experience Recommendations

### 1. **Progressive Disclosure**
- Show basic features first (database, framework, auth)
- Expand advanced features (AI, gRPC, etc.) in collapsible sections
- Use tooltips to explain what each feature does

### 2. **Visual Conflict Prevention**
```css
/* Disable conflicting options */
.feature-option.disabled {
  opacity: 0.5;
  pointer-events: none;
}

/* Highlight conflicts */
.feature-option.conflicted {
  border: 2px solid #ff6b6b;
  background-color: #ffe0e0;
}

/* Show validation states */
.validation-error {
  background: #ffe0e0;
  border-left: 4px solid #ff6b6b;
  padding: 12px;
  margin: 8px 0;
}

.validation-warning {
  background: #fff3cd;
  border-left: 4px solid #ffc107;
  padding: 12px;
  margin: 8px 0;
}

.validation-success {
  background: #d4edda;
  border-left: 4px solid #28a745;
  padding: 12px;
  margin: 8px 0;
}
```

### 3. **Smart Defaults and Suggestions**
```javascript
// Auto-select complementary features
const handleFeatureToggle = (feature) => {
  let newFeatures = [...selectedFeatures];
  
  if (selectedFeatures.includes(feature)) {
    // Remove feature
    newFeatures = newFeatures.filter(f => f !== feature);
  } else {
    // Add feature
    newFeatures.push(feature);
    
    // Smart suggestions
    if (feature === 'auth' && !newFeatures.includes('gin') && !newFeatures.includes('echo')) {
      // Suggest adding a web framework for auth
      showSuggestion('Auth works best with a web framework. Add Gin?', () => {
        setSelectedFeatures([...newFeatures, 'gin']);
      });
    }
    
    if (feature === 'grpc' && !newFeatures.includes('proto')) {
      // Auto-add proto for gRPC (will be done by backend, but show preview)
      showInfo('Protocol Buffers will be automatically added for gRPC');
    }
  }
  
  setSelectedFeatures(newFeatures);
};
```

## 📱 Mobile-Friendly Considerations

### 1. **Compact Error Display**
```javascript
// Mobile-optimized error messages
const MobileValidationFeedback = ({ result }) => (
  <div className="mobile-validation">
    {result.errors?.map((error, index) => (
      <div key={index} className="mobile-error">
        <div className="error-header" onClick={() => toggleExpanded(index)}>
          ❌ {error.message}
          <span className="expand-icon">▼</span>
        </div>
        {expanded[index] && (
          <div className="error-details">
            {error.suggestions.map((suggestion, i) => (
              <div key={i}>💡 {suggestion}</div>
            ))}
          </div>
        )}
      </div>
    ))}
  </div>
);
```

## 🔄 Integration Flow

1. **Page Load**: Fetch `/features` to build UI
2. **Feature Selection**: User selects/deselects features
3. **Real-time Validation**: Call `/generate/validate` on changes
4. **Show Feedback**: Display errors, warnings, auto-additions
5. **Generate**: Only allow if `is_valid: true`

## 🚀 Example Error Messages for UI

```javascript
const ERROR_MESSAGES = {
  'multiple_relational_dbs': 'You can\'t have multiple relational databases in one project. Choose MySQL, PostgreSQL, or SQLite.',
  'multiple_frameworks': 'Choose only one web framework for your project (Gin, Echo, or Fiber).',
  'multiple_orms': 'You can only use one ORM. Choose either GORM or SQLC, not both.',
  'grpc_needs_proto': 'gRPC requires Protocol Buffers. It will be automatically added.',
  'auth_needs_framework': 'Authentication works best with a web framework. Consider adding Gin or Echo.'
};
```

This validation system will make your frontend much more user-friendly by preventing invalid configurations and guiding users toward valid feature combinations! 🎉