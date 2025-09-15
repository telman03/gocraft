# Project History Feature - Frontend Integration Guide

## 🎯 Overview

The backend now includes a comprehensive **Project History System** that allows users to view, manage, and interact with their previously generated projects. This guide provides all the information needed for frontend implementation.

## 🔌 New API Endpoints

### Authentication Required
All history endpoints require JWT authentication via `Authorization: Bearer <token>` header.

---

## 📋 Project History Management

### 1. Get User's Project History
```http
GET /api/history
Authorization: Bearer <jwt_token>
```

**Query Parameters:**
- `page` (int, optional): Page number (default: 1)
- `page_size` (int, optional): Items per page (default: 10, max: 100)
- `search` (string, optional): Search in project name, framework, or features
- `framework` (string, optional): Filter by framework (gin, echo, fiber)
- `frameworks` (string, optional): Filter by multiple frameworks (comma-separated)
- `features` (string, optional): Filter by features (comma-separated)
- `status` (string, optional): Filter by ZIP file status (available, expired, deleted, error)
- `date_from` (string, optional): Filter from date (YYYY-MM-DD format)
- `date_to` (string, optional): Filter to date (YYYY-MM-DD format)
- `sort_by` (string, optional): Sort field (created_at, project_name, framework, zip_file_size, generation_duration_ms)
- `sort_order` (string, optional): Sort order (asc, desc, default: desc)

**Example Request:**
```http
GET /api/history?page=1&page_size=10&framework=gin&search=api&sort_by=created_at&sort_order=desc
```

**Success Response (200):**
```json
{
  "projects": [
    {
      "id": 123,
      "project_name": "my-awesome-api",
      "framework": "gin",
      "features": ["gin", "postgresql", "auth", "redis"],
      "adjusted_features": ["gin", "postgresql", "auth", "redis", "env", "gitignore", "main", "middleware"],
      "zip_file_size": 1048576,
      "zip_file_status": "available",
      "generation_duration_ms": 2500,
      "created_at": "2024-01-15T10:30:00Z",
      "can_download": true,
      "can_regenerate": true
    }
  ],
  "total": 25,
  "page": 1,
  "page_size": 10,
  "total_pages": 3
}
```

**Use Cases:**
- Main project history dashboard
- Search and filter functionality
- Pagination for large project lists

---

### 2. Get Specific Project Details
```http
GET /api/history/{id}
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
```json
{
  "id": 123,
  "project_name": "my-awesome-api",
  "framework": "gin",
  "features": ["gin", "postgresql", "auth", "redis"],
  "adjusted_features": ["gin", "postgresql", "auth", "redis", "env", "gitignore", "main", "middleware"],
  "zip_file_size": 1048576,
  "zip_file_status": "available",
  "generation_duration_ms": 2500,
  "created_at": "2024-01-15T10:30:00Z",
  "can_download": true,
  "can_regenerate": true
}
```

**Error Responses:**
- `400`: Invalid project ID
- `404`: Project not found or access denied
- `401`: Unauthorized

**Use Cases:**
- Project detail modal/page
- Pre-loading data for actions

---

### 3. Delete Project from History
```http
DELETE /api/history/{id}
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
```json
{
  "message": "Project deleted successfully",
  "project_id": 123,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400`: Invalid project ID
- `404`: Project not found or access denied
- `401`: Unauthorized
- `500`: Internal server error

**Use Cases:**
- Delete button with confirmation dialog
- Bulk delete operations

---

## 📥 File Operations

### 4. Download Project ZIP File
```http
GET /api/history/{id}/download
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
- Content-Type: `application/zip`
- Content-Disposition: `attachment; filename="project-name_framework.zip"`
- Returns the actual ZIP file for download

**Error Responses:**
- `400`: Invalid project ID or file path
- `404`: Project not found or file not available
- `410`: File has expired (can regenerate)
- `401`: Unauthorized
- `500`: Internal server error

**Special Error Response for Expired Files (410):**
```json
{
  "error": "Project file has expired",
  "code": "FILE_EXPIRED",
  "details": "File has exceeded the retention period",
  "can_regenerate": "true"
}
```

**Use Cases:**
- Download button functionality
- Automatic file downloads
- Error handling for expired files

---

### 5. Regenerate Project with Same Configuration
```http
POST /api/history/{id}/regenerate
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
- Content-Type: `application/zip`
- Content-Disposition: `attachment; filename="project-name_regenerated_timestamp_framework.zip"`
- Returns the newly generated ZIP file
- Automatically creates new history entry

**Error Responses:**
- `400`: Invalid project ID or insufficient configuration data
- `404`: Project not found or access denied
- `401`: Unauthorized
- `500`: Internal server error

**Use Cases:**
- Regenerate button for expired files
- Quick re-creation of projects
- Automatic download after regeneration

---

## 🔄 Project Duplication

### 6. Duplicate Project Configuration
```http
POST /api/history/duplicate
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "original_project_id": 123,
  "new_project_name": "my-new-project"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Project configuration duplicated successfully",
  "duplicate_config": {
    "project_name": "my-new-project",
    "framework": "gin",
    "features": ["gin", "postgresql", "auth", "redis"],
    "adjusted_features": ["gin", "postgresql", "auth", "redis", "env", "gitignore", "main", "middleware"],
    "original_project": {
      "id": 123,
      "name": "my-awesome-api",
      "created_at": "2024-01-15T10:30:00Z"
    }
  },
  "form_data": {
    "projectName": "my-new-project",
    "framework": "gin",
    "features": ["gin", "postgresql", "auth", "redis"]
  },
  "suggestions": {
    "alternative_names": ["my-awesome-api-copy", "my-awesome-api-v2", "my-awesome-api-20240115"]
  }
}
```

**Error Responses:**
- `400`: Invalid request format or parameters
- `404`: Original project not found or access denied
- `401`: Unauthorized
- `500`: Internal server error

**Use Cases:**
- Duplicate button functionality
- Pre-populate project generator form
- Project name suggestions

---

## 📊 Statistics and Analytics

### 7. Get User's Project Statistics
```http
GET /api/history/stats
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
```json
{
  "total_projects": 20,
  "most_used_framework": "gin",
  "most_used_features": ["gorm", "auth", "redis"],
  "framework_distribution": {
    "gin": 12,
    "echo": 5,
    "fiber": 3
  },
  "recent_activity": [
    {
      "date": "2024-01-15",
      "count": 3
    },
    {
      "date": "2024-01-14",
      "count": 1
    }
  ]
}
```

**Use Cases:**
- Dashboard statistics cards
- Analytics charts and graphs
- User insights and patterns

---

### 8. Get Dashboard Data (Optimized)
```http
GET /api/history/dashboard
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
```json
{
  "summary": {
    "total_projects": 20,
    "recent_projects": 5,
    "available_downloads": 15,
    "expired_files": 3
  },
  "recent_projects": [
    {
      "id": 123,
      "project_name": "my-awesome-api",
      "framework": "gin",
      "created_at": "2024-01-15T10:30:00Z",
      "zip_file_status": "available"
    }
  ],
  "statistics": {
    "most_used_framework": "gin",
    "framework_distribution": {
      "gin": 12,
      "echo": 5,
      "fiber": 3
    },
    "recent_activity": [
      {
        "date": "2024-01-15",
        "count": 3
      }
    ]
  }
}
```

**Use Cases:**
- Main dashboard overview
- Quick access to recent projects
- Performance-optimized data loading

---

### 9. Get Cache Performance Information
```http
GET /api/history/cache-info
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
```json
{
  "cache_stats": {
    "hit_rate": 0.85,
    "total_requests": 1000,
    "cache_hits": 850,
    "cache_misses": 150,
    "cache_size": 50
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Use Cases:**
- Admin/debug information
- Performance monitoring
- Cache optimization insights

---

## 🎨 Frontend Implementation Guidelines

### 1. **Project History Dashboard Layout**

```javascript
// Recommended component structure
const ProjectHistoryDashboard = () => {
  return (
    <div className="dashboard">
      <Header />
      <div className="dashboard-content">
        <Sidebar>
          <StatsOverview />
          <QuickActions />
        </Sidebar>
        <MainContent>
          <SearchAndFilters />
          <ProjectList />
          <Pagination />
        </MainContent>
      </div>
    </div>
  );
};
```

### 2. **Project Card Component**

```javascript
// Individual project display
const ProjectCard = ({ project }) => {
  return (
    <div className="project-card">
      <div className="project-header">
        <h3>{project.project_name}</h3>
        <span className="framework-badge">{project.framework}</span>
      </div>
      
      <div className="project-details">
        <div className="features">
          {project.features.map(feature => (
            <span key={feature} className="feature-tag">{feature}</span>
          ))}
        </div>
        
        <div className="metadata">
          <span>Created: {formatDate(project.created_at)}</span>
          <span>Size: {formatFileSize(project.zip_file_size)}</span>
          <span className={`status ${project.zip_file_status}`}>
            {project.zip_file_status}
          </span>
        </div>
      </div>
      
      <div className="project-actions">
        <button 
          disabled={!project.can_download}
          onClick={() => downloadProject(project.id)}
        >
          Download
        </button>
        
        <button onClick={() => duplicateProject(project.id)}>
          Duplicate
        </button>
        
        <button 
          disabled={!project.can_regenerate}
          onClick={() => regenerateProject(project.id)}
        >
          Regenerate
        </button>
        
        <button 
          className="danger"
          onClick={() => deleteProject(project.id)}
        >
          Delete
        </button>
      </div>
    </div>
  );
};
```

### 3. **Search and Filter Implementation**

```javascript
const SearchAndFilters = ({ onFiltersChange }) => {
  const [filters, setFilters] = useState({
    search: '',
    framework: '',
    status: '',
    dateFrom: '',
    dateTo: '',
    sortBy: 'created_at',
    sortOrder: 'desc'
  });

  const handleFilterChange = (key, value) => {
    const newFilters = { ...filters, [key]: value };
    setFilters(newFilters);
    onFiltersChange(newFilters);
  };

  return (
    <div className="search-filters">
      <input
        type="text"
        placeholder="Search projects..."
        value={filters.search}
        onChange={(e) => handleFilterChange('search', e.target.value)}
      />
      
      <select
        value={filters.framework}
        onChange={(e) => handleFilterChange('framework', e.target.value)}
      >
        <option value="">All Frameworks</option>
        <option value="gin">Gin</option>
        <option value="echo">Echo</option>
        <option value="fiber">Fiber</option>
      </select>
      
      <select
        value={filters.status}
        onChange={(e) => handleFilterChange('status', e.target.value)}
      >
        <option value="">All Status</option>
        <option value="available">Available</option>
        <option value="expired">Expired</option>
        <option value="deleted">Deleted</option>
      </select>
      
      <input
        type="date"
        placeholder="From Date"
        value={filters.dateFrom}
        onChange={(e) => handleFilterChange('dateFrom', e.target.value)}
      />
      
      <input
        type="date"
        placeholder="To Date"
        value={filters.dateTo}
        onChange={(e) => handleFilterChange('dateTo', e.target.value)}
      />
    </div>
  );
};
```

### 4. **API Client Service**

```javascript
// API service for project history
class ProjectHistoryAPI {
  constructor(baseURL, authToken) {
    this.baseURL = baseURL;
    this.authToken = authToken;
  }

  async getProjects(filters = {}) {
    const params = new URLSearchParams(filters);
    const response = await fetch(`${this.baseURL}/api/history?${params}`, {
      headers: {
        'Authorization': `Bearer ${this.authToken}`,
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to fetch projects: ${response.statusText}`);
    }
    
    return response.json();
  }

  async getProject(id) {
    const response = await fetch(`${this.baseURL}/api/history/${id}`, {
      headers: {
        'Authorization': `Bearer ${this.authToken}`
      }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to fetch project: ${response.statusText}`);
    }
    
    return response.json();
  }

  async downloadProject(id) {
    const response = await fetch(`${this.baseURL}/api/history/${id}/download`, {
      headers: {
        'Authorization': `Bearer ${this.authToken}`
      }
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Download failed');
    }
    
    // Handle file download
    const blob = await response.blob();
    const filename = response.headers.get('Content-Disposition')
      ?.split('filename=')[1]?.replace(/"/g, '') || 'project.zip';
    
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    window.URL.revokeObjectURL(url);
  }

  async regenerateProject(id) {
    const response = await fetch(`${this.baseURL}/api/history/${id}/regenerate`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.authToken}`
      }
    });
    
    if (!response.ok) {
      throw new Error('Regeneration failed');
    }
    
    // Handle file download (same as downloadProject)
    const blob = await response.blob();
    const filename = response.headers.get('Content-Disposition')
      ?.split('filename=')[1]?.replace(/"/g, '') || 'project.zip';
    
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    window.URL.revokeObjectURL(url);
  }

  async duplicateProject(originalId, newName) {
    const response = await fetch(`${this.baseURL}/api/history/duplicate`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.authToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        original_project_id: originalId,
        new_project_name: newName
      })
    });
    
    if (!response.ok) {
      throw new Error('Duplication failed');
    }
    
    return response.json();
  }

  async deleteProject(id) {
    const response = await fetch(`${this.baseURL}/api/history/${id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${this.authToken}`
      }
    });
    
    if (!response.ok) {
      throw new Error('Deletion failed');
    }
    
    return response.json();
  }

  async getStats() {
    const response = await fetch(`${this.baseURL}/api/history/stats`, {
      headers: {
        'Authorization': `Bearer ${this.authToken}`
      }
    });
    
    if (!response.ok) {
      throw new Error('Failed to fetch statistics');
    }
    
    return response.json();
  }

  async getDashboardData() {
    const response = await fetch(`${this.baseURL}/api/history/dashboard`, {
      headers: {
        'Authorization': `Bearer ${this.authToken}`
      }
    });
    
    if (!response.ok) {
      throw new Error('Failed to fetch dashboard data');
    }
    
    return response.json();
  }
}
```

### 5. **Error Handling**

```javascript
// Error handling utility
const handleAPIError = (error, showToast) => {
  console.error('API Error:', error);
  
  if (error.message.includes('401')) {
    // Redirect to login
    window.location.href = '/login';
    return;
  }
  
  if (error.message.includes('404')) {
    showToast('Project not found or access denied', 'error');
    return;
  }
  
  if (error.message.includes('410')) {
    showToast('File has expired. You can regenerate the project.', 'warning');
    return;
  }
  
  showToast('An error occurred. Please try again.', 'error');
};
```

### 6. **Statistics Dashboard**

```javascript
const StatsDashboard = () => {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await api.getStats();
        setStats(data);
      } catch (error) {
        handleAPIError(error, showToast);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  if (loading) return <LoadingSpinner />;

  return (
    <div className="stats-dashboard">
      <div className="stats-cards">
        <StatCard
          title="Total Projects"
          value={stats.total_projects}
          icon="📊"
        />
        <StatCard
          title="Most Used Framework"
          value={stats.most_used_framework}
          icon="🚀"
        />
        <StatCard
          title="Top Features"
          value={stats.most_used_features.slice(0, 3).join(', ')}
          icon="⭐"
        />
      </div>
      
      <div className="charts">
        <FrameworkDistributionChart data={stats.framework_distribution} />
        <ActivityChart data={stats.recent_activity} />
      </div>
    </div>
  );
};
```

## 🔐 Security Considerations

### 1. **Authentication**
- All endpoints require JWT token in Authorization header
- Handle 401 responses by redirecting to login
- Implement token refresh mechanism

### 2. **Data Validation**
- Validate all user inputs before sending to API
- Sanitize project names and search queries
- Implement client-side validation for better UX

### 3. **File Downloads**
- Handle download errors gracefully
- Show progress indicators for large files
- Implement download retry mechanism

## 🎯 User Experience Guidelines

### 1. **Loading States**
- Show skeleton loaders for project lists
- Display progress indicators for downloads
- Use optimistic updates for quick actions

### 2. **Error Feedback**
- Show clear error messages for failed operations
- Provide actionable suggestions (e.g., "Regenerate" for expired files)
- Use toast notifications for non-blocking feedback

### 3. **Responsive Design**
- Implement mobile-friendly project cards
- Use responsive tables for project lists
- Optimize touch targets for mobile devices

### 4. **Performance**
- Implement virtual scrolling for large lists
- Use pagination to limit data loading
- Cache frequently accessed data

## 📱 Mobile Considerations

### 1. **Touch-Friendly Interface**
- Larger buttons and touch targets
- Swipe gestures for project actions
- Mobile-optimized modals and dialogs

### 2. **Responsive Layout**
- Stack project cards vertically on mobile
- Collapsible filters and search
- Bottom sheet for project actions

This comprehensive guide provides everything needed to implement the project history feature on the frontend. The backend is fully implemented and ready to support all these use cases! 🚀