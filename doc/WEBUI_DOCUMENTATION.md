# Buck It Up - Web UI Documentation

## Overview

The Buck It Up server now includes a modern web-based admin dashboard accessible at `/ui`. This provides a graphical interface for managing buckets and objects without needing to use curl or PowerShell commands.

## Features

- **Admin Authentication** - Secure login using the admin password
- **Bucket Management** - View, create, and delete buckets
- **Object Management** - Upload, view, download, and delete objects
- **File Upload** - Drag-and-drop file upload or paste text content
- **Content Preview** - Preview text files directly in the browser
- **Responsive Design** - Works on desktop and mobile devices

## Getting Started

### 1. Set the Admin Password

Before accessing the UI, make sure you have set the `BUCKITUP_ADMIN_PASSWORD` environment variable:

**Windows PowerShell:**
```powershell
$env:BUCKITUP_ADMIN_PASSWORD="your-secure-admin-password"
.\buck_It_Up.exe
```

**Linux/Mac:**
```bash
export BUCKITUP_ADMIN_PASSWORD="your-secure-admin-password"
./buck_It_Up
```

### 2. Access the Web UI

Open your web browser and navigate to:

```
http://localhost:8080/ui
```

This will redirect you to the login page.

### 3. Login

Enter your admin password (the same one you set in `BUCKITUP_ADMIN_PASSWORD`) and click "Login".

## Using the Dashboard

### Main Dashboard

After logging in, you'll see the main dashboard with:

- **List of all buckets** - Shows all existing buckets with their IDs and creation dates
- **Create Bucket button** - Click to create a new bucket
- **View Objects button** - Click on any bucket to see its contents
- **Delete button** - Remove a bucket (must be empty)

### Creating a Bucket

1. Click the **"+ Create Bucket"** button
2. Enter a bucket name (letters, numbers, hyphens, and underscores only)
3. Click **"Create"**
4. The bucket will be created with three access keys (read-only, read-write, and all permissions)

### Viewing Bucket Contents

1. Click **"View Objects"** on any bucket card
2. You'll see a table with all objects in the bucket showing:
   - Object key (name)
   - File size
   - Content type
   - Creation date
   - Action buttons

### Uploading Objects

1. Navigate to a bucket's detail view
2. Click **"+ Upload Object"**
3. Enter an object key (can include slashes for folder structure, e.g., `images/photo.jpg`)
4. Choose upload method:
   - **File Upload**: Click to browse or drag-and-drop a file
   - **Text Content**: Paste or type text content directly
5. Optionally specify a content type (auto-detected for files)
6. Click **"Upload"**

### Viewing Objects

1. In a bucket's detail view, click **"View"** on any object
2. You'll see:
   - Object metadata (content type, size)
   - Content preview (for text files)
   - Download button

### Downloading Objects

- Click **"Download"** on any object in the table
- Or click **"Download"** in the object preview modal
- The file will be saved to your browser's download folder

### Deleting Objects

1. Click **"Delete"** on any object
2. Confirm the deletion
3. The object will be permanently removed

### Deleting Buckets

1. From the main dashboard, click **"Delete"** on a bucket card
2. Confirm the deletion
3. Note: The bucket must be empty (delete all objects first)

## Routes

The following routes are available:

- `/ui` or `/ui/` - Redirects to login page
- `/ui/login` - Admin login page
- `/ui/dashboard` - Main dashboard (shows all buckets)
- `/ui/bucket/{bucketName}` - Bucket detail view (shows all objects)

## Security Features

- **Client-side authentication** - Admin credentials are stored in browser's localStorage
- **Automatic session management** - Invalid credentials redirect to login
- **Server-side validation** - All API calls require valid admin authentication
- **HTTPS recommended** - Use HTTPS in production to encrypt credentials in transit

## Browser Compatibility

The UI works with all modern browsers:

- Chrome/Edge (recommended)
- Firefox
- Safari
- Opera

JavaScript must be enabled for the UI to function.

## Troubleshooting

### "Invalid admin password" error

- Make sure the `BUCKITUP_ADMIN_PASSWORD` environment variable is set before starting the server
- Verify you're entering the correct password
- The password is case-sensitive

### Can't access the UI

- Make sure the server is running on port 8080 (or check your `PORT` environment variable)
- Check that your firewall isn't blocking the connection
- Try accessing `http://localhost:8080/health` to verify the server is running

### Objects not displaying

- Check browser console for errors (F12 → Console)
- Verify the bucket exists and contains objects
- Try refreshing the page

### Upload fails

- Check file size (very large files may cause issues)
- Verify the object key doesn't contain invalid characters
- Check browser console for detailed error messages

## API Integration

The UI uses the same REST API as the command-line tools. All actions performed in the UI correspond to API calls:

| UI Action | API Call |
|-----------|----------|
| Login | `LIST /` (validates credentials) |
| List buckets | `LIST /` |
| Create bucket | `POST /` |
| Delete bucket | `DELETE /{name}` |
| List objects | `LIST /{bucketName}` |
| Upload object | `POST /{bucketName}/upload` |
| View object | `GET /{bucketName}/all/{key}` |
| Download object | `GET /{bucketName}/all/{key}` |
| Delete object | `DELETE /{bucketName}/{key}` |

## Logout

Click the **"Logout"** button in the top-right corner to clear your session and return to the login page.

## Tips

1. **Object Keys**: Use forward slashes in object keys to create a folder-like structure (e.g., `docs/readme.txt`)
2. **Content Types**: Specifying the correct content type helps with browser preview and download behavior
3. **Text Files**: Text files (text/*, application/json, etc.) can be previewed directly in the UI
4. **Binary Files**: Binary files show a size indicator and must be downloaded to view
5. **Bulk Operations**: For bulk uploads or complex operations, consider using the API directly

## Screenshots Description

### Login Page
- Clean, centered login form with gradient background
- Single password field for admin authentication
- Error messages displayed inline

### Dashboard
- Grid layout of bucket cards
- Each card shows bucket name, ID, and creation date
- Quick action buttons for viewing and deleting
- Create new bucket modal dialog

### Bucket View
- Breadcrumb navigation back to dashboard
- Table view of all objects with sortable columns
- Upload modal with file picker or text input
- Object preview modal with download option

