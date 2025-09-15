# Requirements Document

## Introduction

The project history feature will allow users to view and manage a history of all their previously generated projects. This feature enhances the user experience by providing visibility into past project generations, enabling users to track their development patterns, re-download previous projects, and maintain a record of their GoCraft usage.

## Requirements

### Requirement 1

**User Story:** As a registered user, I want to see a history of all my generated projects, so that I can track what I've built and easily access previous configurations.

#### Acceptance Criteria

1. WHEN a user generates a project THEN the system SHALL automatically save the project details to their history
2. WHEN a user accesses their dashboard THEN the system SHALL display a list of their project history
3. WHEN displaying project history THEN the system SHALL show project name, framework, features, generation date, and download status
4. WHEN a user views their project history THEN the system SHALL display projects in reverse chronological order (newest first)
5. WHEN a user has no project history THEN the system SHALL display an appropriate empty state message

### Requirement 2

**User Story:** As a registered user, I want to see detailed information about each generated project, so that I can understand what features and configurations I used.

#### Acceptance Criteria

1. WHEN displaying a project in history THEN the system SHALL show the project name
2. WHEN displaying a project in history THEN the system SHALL show the selected framework (Gin, Echo, or Fiber)
3. WHEN displaying a project in history THEN the system SHALL show all selected features as a readable list
4. WHEN displaying a project in history THEN the system SHALL show the generation timestamp
5. WHEN displaying a project in history THEN the system SHALL show the file size of the generated ZIP
6. WHEN displaying a project in history THEN the system SHALL indicate if the ZIP file is still available for download

### Requirement 3

**User Story:** As a registered user, I want to re-download previously generated projects, so that I can access my past work without regenerating it.

#### Acceptance Criteria

1. WHEN a project ZIP file is still available THEN the system SHALL provide a download link
2. WHEN a user clicks a download link THEN the system SHALL serve the original ZIP file
3. WHEN a project ZIP file is no longer available THEN the system SHALL show "File Expired" status
4. WHEN a project ZIP file is no longer available THEN the system SHALL offer a "Regenerate" option
5. WHEN a user chooses to regenerate THEN the system SHALL use the same configuration to create a new project

### Requirement 4

**User Story:** As a registered user, I want to duplicate a previous project configuration, so that I can create similar projects with minor modifications.

#### Acceptance Criteria

1. WHEN viewing project history THEN the system SHALL provide a "Duplicate" action for each project
2. WHEN a user clicks "Duplicate" THEN the system SHALL pre-populate the project generator with the same framework and features
3. WHEN duplicating a project THEN the system SHALL suggest a new project name based on the original
4. WHEN duplicating a project THEN the system SHALL allow the user to modify any settings before generating
5. WHEN a duplicated project is generated THEN the system SHALL save it as a new entry in the history

### Requirement 5

**User Story:** As a registered user, I want to delete projects from my history, so that I can manage my project list and remove unwanted entries.

#### Acceptance Criteria

1. WHEN viewing project history THEN the system SHALL provide a "Delete" action for each project
2. WHEN a user clicks "Delete" THEN the system SHALL show a confirmation dialog
3. WHEN a user confirms deletion THEN the system SHALL remove the project from their history
4. WHEN a project is deleted THEN the system SHALL also remove the associated ZIP file if it exists
5. WHEN a project is deleted THEN the system SHALL show a success message and update the history list

### Requirement 6

**User Story:** As a registered user, I want to search and filter my project history, so that I can quickly find specific projects.

#### Acceptance Criteria

1. WHEN viewing project history THEN the system SHALL provide a search input field
2. WHEN a user enters search terms THEN the system SHALL filter projects by name, framework, or features
3. WHEN viewing project history THEN the system SHALL provide filter options for framework type
4. WHEN viewing project history THEN the system SHALL provide filter options for date ranges
5. WHEN filters are applied THEN the system SHALL update the displayed list in real-time
6. WHEN no projects match the search/filter criteria THEN the system SHALL show an appropriate "no results" message

### Requirement 7

**User Story:** As a registered user, I want to see statistics about my project generation patterns, so that I can understand my development preferences.

#### Acceptance Criteria

1. WHEN viewing the dashboard THEN the system SHALL display total number of projects generated
2. WHEN viewing the dashboard THEN the system SHALL display most frequently used framework
3. WHEN viewing the dashboard THEN the system SHALL display most frequently used features
4. WHEN viewing the dashboard THEN the system SHALL display generation activity over time (last 30 days)
5. WHEN a user has insufficient data THEN the system SHALL show appropriate placeholder messages

### Requirement 8

**User Story:** As a system administrator, I want project history data to be properly managed, so that the system remains performant and storage is optimized.

#### Acceptance Criteria

1. WHEN a project is generated THEN the system SHALL store only essential metadata in the database
2. WHEN ZIP files are older than 30 days THEN the system SHALL automatically clean them up
3. WHEN ZIP files are cleaned up THEN the system SHALL update the project status to "File Expired"
4. WHEN the database grows large THEN the system SHALL maintain good query performance through proper indexing
5. WHEN a user account is deleted THEN the system SHALL cascade delete all associated project history

### Requirement 9

**User Story:** As a registered user, I want my project history to be secure and private, so that only I can access my project information.

#### Acceptance Criteria

1. WHEN accessing project history THEN the system SHALL require valid authentication
2. WHEN a user requests project history THEN the system SHALL only return projects belonging to that user
3. WHEN downloading a project ZIP THEN the system SHALL verify the user owns that project
4. WHEN accessing project history via API THEN the system SHALL validate JWT tokens
5. WHEN unauthorized access is attempted THEN the system SHALL return appropriate error responses

### Requirement 10

**User Story:** As a registered user, I want the project history feature to work seamlessly with the existing project generation flow, so that my experience is smooth and intuitive.

#### Acceptance Criteria

1. WHEN a project generation completes successfully THEN the system SHALL automatically add it to history without user intervention
2. WHEN a project generation fails THEN the system SHALL NOT add an entry to the history
3. WHEN viewing project history THEN the system SHALL maintain the same UI/UX patterns as the rest of the application
4. WHEN the history feature is unavailable THEN the system SHALL gracefully degrade and still allow project generation
5. WHEN API responses include history data THEN the system SHALL follow the established response format patterns