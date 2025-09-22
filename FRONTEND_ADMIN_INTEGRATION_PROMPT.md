# 🚀 Admin Dashboard API Integration - Replace Mock Data with Real APIs

## 📋 Task Overview

The backend now provides **real admin API endpoints** with proper authentication and role-based access control. Your task is to **replace all mock data** in the admin dashboard with actual API calls to these endpoints.

## 🔐 Authentication Changes

### **CRITICAL: Admin Role Required**
- All admin endpoints now require **admin role** (not just authentication)
- Regular users will get `403 Forbidden` with `ADMIN_REQUIRED` error
- You need to handle role-based access in the frontend

### **Admin Credentials for Testing:**
```
Email: admin@gocraft.dev
Password: AdminPassword123!
Role: admin
```

## 🔌 Real API Endpoints to Integrate

### 1. **Admin User Management**
```http
GET /api/admin/users?page=1&page_size=20
Authorization: Bearer <admin_token>
```

**Real Response:**
```json
{
  "users": [
    {
      "id": 20,
      "email": "admin@gocraft.dev",
      "role": "admin",
      "is_verified": true,
      "created_at": "2025-09-22T15:55:28.471059+04:00",
      "updated_at": "2025-09-22T15:55:28.471059+04:00"
    },
    {
      "id": 18,
      "email": "telmangadimov1@gmail.com", 
      "role": "user",
      "is_verified": true,
      "created_at": "2025-09-15T16:01:23.483412+04:00",
      "updated_at": "2025-09-15T16:01:48.026454+04:00"
    }
  ],
  "total": 5,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

**Replace Mock Data:** Update `UserManagement` component to use real user data

---

### 2. **Admin Statistics Dashboard**
```http
GET /api/admin/stats
Authorization: Bearer <admin_token>
```

**Real Response:**
```json
{
  "total_users": 5,
  "verified_users": 4,
  "admin_users": 1,
  "regular_users": 4,
  "total_projects": 3,
  "recent_users": 5,
  "verification_rate": 80.0
}
```

**Replace Mock Data:** Update overview stats cards with real metrics

---

### 3. **Update User Role**
```http
PUT /api/admin/users/{id}/role
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "role": "admin"  // or "user"
}
```

**Real Response:**
```json
{
  "message": "User role updated successfully",
  "user_id": 18,
  "email": "user@example.com",
  "old_role": "user",
  "new_role": "admin"
}
```

**New Feature:** Add role management functionality to user table

---

### 4. **Project History (Existing)**
```http
GET /api/history?page=1&page_size=10
Authorization: Bearer <admin_token>
```

**Real Response:**
```json
{
  "projects": [
    {
      "id": 2,
      "project_name": "test-api",
      "framework": "gin", 
      "features": ["gin", "postgresql", "auth"],
      "zip_file_size": 8196,
      "zip_file_status": "available",
      "created_at": "2025-09-15T16:15:08.909409+04:00",
      "can_download": true,
      "can_regenerate": true
    }
  ],
  "total": 2,
  "page": 1,
  "page_size": 10,
  "total_pages": 1
}
```

**Replace Mock Data:** Update projects tab with real project data

---

### 5. **System Maintenance (Existing)**
```http
GET /api/maintenance/status
Authorization: Bearer <admin_token>
```

**Real Response:**
```json
{
  "database_maintenance": {
    "running": true,
    "config": {
      "maintenance_interval": 86400000000000,
      "cleanup_batch_size": 1000,
      "enable_scheduling": true
    }
  },
  "file_cleanup": {
    "running": true,
    "config": {
      "cleanup_interval": 86400000000000,
      "batch_size": 100,
      "retention_period": 2592000000000000
    }
  }
}
```

**Replace Mock Data:** Update system monitoring with real service status

---

## 🎯 Frontend Integration Tasks

### **Task 1: Update API Client**
Replace mock data calls with real API endpoints:

```javascript
// ❌ OLD: Mock data
const mockUsers = [
  { id: 1, name: "John Doe", email: "john@example.com", status: "active" }
];

// ✅ NEW: Real API call
const getUsers = async (page = 1, pageSize = 20) => {
  const response = await fetch(`/api/admin/users?page=${page}&page_size=${pageSize}`, {
    headers: {
      'Authorization': `Bearer ${getAdminToken()}`,
      'Content-Type': 'application/json'
    }
  });
  
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Admin access required');
    }
    throw new Error('Failed to fetch users');
  }
  
  return response.json();
};
```

### **Task 2: Add Role Management**
Add role update functionality to user management:

```javascript
const updateUserRole = async (userId, newRole) => {
  const response = await fetch(`/api/admin/users/${userId}/role`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${getAdminToken()}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ role: newRole })
  });
  
  if (!response.ok) {
    throw new Error('Failed to update user role');
  }
  
  return response.json();
};
```

### **Task 3: Update Stats Cards**
Replace mock statistics with real data:

```javascript
// ❌ OLD: Mock stats
const mockStats = {
  totalUsers: 12543,
  activeUsers: 8921,
  totalProjects: 45678,
  serverUptime: 99.9
};

// ✅ NEW: Real API data
const loadAdminStats = async () => {
  const response = await fetch('/api/admin/stats', {
    headers: {
      'Authorization': `Bearer ${getAdminToken()}`
    }
  });
  
  const stats = await response.json();
  
  return {
    totalUsers: stats.total_users,
    verifiedUsers: stats.verified_users,
    adminUsers: stats.admin_users,
    totalProjects: stats.total_projects,
    verificationRate: stats.verification_rate
  };
};
```

### **Task 4: Add Error Handling**
Handle admin-specific errors:

```javascript
const handleAdminError = (error, response) => {
  if (response?.status === 403) {
    // User doesn't have admin role
    showError('Admin access required. Please contact an administrator.');
    redirectToLogin();
    return;
  }
  
  if (response?.status === 401) {
    // Token expired or invalid
    showError('Session expired. Please log in again.');
    redirectToLogin();
    return;
  }
  
  // Generic error
  showError(error.message || 'An error occurred');
};
```

### **Task 5: Update User Table**
Add role management to the user table:

```javascript
const UserTable = ({ users, onRoleUpdate }) => {
  const handleRoleChange = async (userId, newRole) => {
    try {
      await updateUserRole(userId, newRole);
      onRoleUpdate(); // Refresh the user list
      showSuccess('User role updated successfully');
    } catch (error) {
      showError('Failed to update user role');
    }
  };

  return (
    <table>
      <thead>
        <tr>
          <th>Email</th>
          <th>Role</th>
          <th>Status</th>
          <th>Created</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {users.map(user => (
          <tr key={user.id}>
            <td>{user.email}</td>
            <td>
              <select 
                value={user.role} 
                onChange={(e) => handleRoleChange(user.id, e.target.value)}
              >
                <option value="user">User</option>
                <option value="admin">Admin</option>
              </select>
            </td>
            <td>{user.is_verified ? 'Verified' : 'Pending'}</td>
            <td>{formatDate(user.created_at)}</td>
            <td>
              <button onClick={() => viewUser(user.id)}>View</button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
};
```

## 🔐 Authentication Integration

### **Admin Token Management**
```javascript
// Check if user has admin role
const checkAdminAccess = async () => {
  try {
    const response = await fetch('/api/admin/stats', {
      headers: {
        'Authorization': `Bearer ${getToken()}`
      }
    });
    
    if (response.status === 403) {
      // User is authenticated but not admin
      return false;
    }
    
    return response.ok;
  } catch (error) {
    return false;
  }
};

// Protect admin routes
const AdminRoute = ({ children }) => {
  const [hasAdminAccess, setHasAdminAccess] = useState(null);
  
  useEffect(() => {
    checkAdminAccess().then(setHasAdminAccess);
  }, []);
  
  if (hasAdminAccess === null) {
    return <LoadingSpinner />;
  }
  
  if (!hasAdminAccess) {
    return <AccessDenied message="Admin access required" />;
  }
  
  return children;
};
```

## 📊 Data Mapping Guide

### **User Data Mapping**
```javascript
// Backend response -> Frontend display
const mapUserData = (backendUser) => ({
  id: backendUser.id,
  name: backendUser.email, // Use email as name
  email: backendUser.email,
  role: backendUser.role, // 'user' or 'admin'
  status: backendUser.is_verified ? 'active' : 'pending',
  joinDate: backendUser.created_at,
  lastLogin: backendUser.updated_at // Approximate
});
```

### **Stats Data Mapping**
```javascript
// Backend response -> Frontend cards
const mapStatsData = (backendStats) => ({
  totalUsers: {
    value: backendStats.total_users,
    label: 'Total Users'
  },
  activeUsers: {
    value: backendStats.verified_users,
    label: 'Verified Users'
  },
  adminUsers: {
    value: backendStats.admin_users,
    label: 'Admin Users'
  },
  verificationRate: {
    value: `${backendStats.verification_rate.toFixed(1)}%`,
    label: 'Verification Rate'
  }
});
```

## 🚨 Error Scenarios to Handle

### **1. Non-Admin User Access**
```json
{
  "success": false,
  "error": {
    "code": "ACCESS_DENIED",
    "message": "Admin access required"
  },
  "context": {
    "code": "ADMIN_REQUIRED",
    "message": "You need admin privileges to access this resource"
  }
}
```

### **2. Invalid User ID**
```json
{
  "error": "User not found",
  "code": "USER_NOT_FOUND"
}
```

### **3. Self-Demotion Prevention**
```json
{
  "error": "Cannot change your own admin role",
  "code": "CANNOT_DEMOTE_SELF"
}
```

## 🎨 UI Updates Needed

### **1. Add Role Badges**
```javascript
const RoleBadge = ({ role }) => (
  <span className={`badge ${role === 'admin' ? 'badge-admin' : 'badge-user'}`}>
    {role.toUpperCase()}
  </span>
);
```

### **2. Add Admin Access Warning**
```javascript
const AdminAccessWarning = () => (
  <div className="alert alert-warning">
    ⚠️ Admin access required. Only administrators can view this page.
  </div>
);
```

### **3. Update Loading States**
```javascript
const AdminDashboard = () => {
  const [loading, setLoading] = useState(true);
  const [hasAccess, setHasAccess] = useState(false);
  
  useEffect(() => {
    checkAdminAccess().then(access => {
      setHasAccess(access);
      setLoading(false);
    });
  }, []);
  
  if (loading) return <LoadingSpinner />;
  if (!hasAccess) return <AdminAccessWarning />;
  
  return <AdminDashboardContent />;
};
```

## ✅ Testing Checklist

- [ ] Replace all mock data with real API calls
- [ ] Add proper error handling for 403/401 responses
- [ ] Implement role management functionality
- [ ] Add admin access checks to protected routes
- [ ] Update user table with real user data
- [ ] Update stats cards with real metrics
- [ ] Test with admin credentials: `admin@gocraft.dev` / `AdminPassword123!`
- [ ] Test access denial with regular user credentials
- [ ] Verify role update functionality works
- [ ] Ensure pagination works with real data

## 🚀 Expected Outcome

After integration, the admin dashboard should:
1. ✅ Use real user data from the database
2. ✅ Show actual system statistics
3. ✅ Allow role management (promote/demote users)
4. ✅ Properly handle admin access control
5. ✅ Display real project data
6. ✅ Show actual system maintenance status

The dashboard will be **fully functional** with real backend data instead of mock data! 🎯