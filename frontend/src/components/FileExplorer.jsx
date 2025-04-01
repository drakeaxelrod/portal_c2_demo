import React, { useState, useEffect, useRef } from 'react';
import './FileExplorer.css';

const API_URL = 'http://localhost:8080';

const FileExplorer = ({ agent, onClose }) => {
  const [currentPath, setCurrentPath] = useState('/');
  const [fileList, setFileList] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedFile, setSelectedFile] = useState(null);
  const [breadcrumbs, setBreadcrumbs] = useState([{ name: 'Root', path: '/' }]);
  const [isDragging, setIsDragging] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [draggedFile, setDraggedFile] = useState(null);
  const [dropTarget, setDropTarget] = useState(null);
  const fileInputRef = useRef(null);
  const [downloadingFile, setDownloadingFile] = useState(null);

  useEffect(() => {
    loadFileList(currentPath);
  }, [agent, currentPath]);

  const loadFileList = async (path) => {
    if (!agent || !agent.id) return;
    
    setIsLoading(true);
    setError(null);
    
    try {
      // Execute an ls command to get directory contents
      const response = await fetch(`${API_URL}/api/agents/${agent.id}/command`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          command: `ls -la "${path}"`,
          type: 'shell',
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch directory contents: ${response.status}`);
      }

      const result = await response.json();
      
      if (!result.success) {
        throw new Error(result.error || 'Failed to list directory contents');
      }

      // Parse the ls output to get files and directories
      const files = parseDirectoryListing(result.result, path);
      setFileList(files);
    } catch (err) {
      console.error('Error loading file list:', err);
      setError(`Failed to load files: ${err.message}`);
    } finally {
      setIsLoading(false);
    }
  };

  const parseDirectoryListing = (output, currentPath) => {
    // Split output by lines
    const lines = output.trim().split('\n');
    
    // Skip the total line and the . and .. entries (first 3 lines usually)
    const fileLines = lines.slice(1);
    
    return fileLines.map(line => {
      // Parse ls -la output format
      // Example: drwxr-xr-x 2 user group 4096 Apr 1 12:34 dirname
      const parts = line.trim().split(/\s+/);
      
      // Need at least 9 parts for a valid ls -la line
      if (parts.length < 9) return null;
      
      const permissions = parts[0];
      const isDirectory = permissions.startsWith('d');
      const isLink = permissions.startsWith('l');
      
      // The filename is everything after the date/time (parts[8] and onwards)
      const nameIndex = 8;
      let name = parts.slice(nameIndex).join(' ');
      
      // Handle symbolic links (remove the -> target part)
      if (isLink && name.includes(' -> ')) {
        name = name.split(' -> ')[0];
      }
      
      // Skip . and .. entries
      if (name === '.' || name === '..') return null;
      
      return {
        name,
        path: `${currentPath === '/' ? '' : currentPath}/${name}`,
        isDirectory,
        isLink,
        size: parseInt(parts[4], 10),
        permissions,
        modifiedTime: `${parts[5]} ${parts[6]} ${parts[7]}`
      };
    }).filter(Boolean); // Remove null entries
  };

  const navigateToDirectory = (path, dirName) => {
    // Update breadcrumbs
    if (path === '/') {
      setBreadcrumbs([{ name: 'Root', path: '/' }]);
    } else {
      const pathParts = path.split('/').filter(Boolean);
      const newBreadcrumbs = [{ name: 'Root', path: '/' }];
      
      let currentPath = '';
      for (const part of pathParts) {
        currentPath += '/' + part;
        newBreadcrumbs.push({
          name: part,
          path: currentPath
        });
      }
      
      setBreadcrumbs(newBreadcrumbs);
    }
    
    setCurrentPath(path);
  };

  const handleFileClick = (file) => {
    setSelectedFile(file);
    
    if (file.isDirectory) {
      navigateToDirectory(file.path, file.name);
    }
  };

  const handleBreadcrumbClick = (breadcrumb) => {
    navigateToDirectory(breadcrumb.path);
  };

  const goToParentDirectory = () => {
    if (currentPath === '/') return;
    
    const pathParts = currentPath.split('/').filter(Boolean);
    pathParts.pop();
    const parentPath = pathParts.length === 0 ? '/' : '/' + pathParts.join('/');
    
    navigateToDirectory(parentPath);
  };

  // Get appropriate icon for file type
  const getFileIcon = (file) => {
    if (file.isDirectory) return '📁';
    if (file.isLink) return '🔗';
    
    // Check extension for common file types
    const extension = file.name.split('.').pop().toLowerCase();
    
    switch(extension) {
      case 'txt':
      case 'md':
      case 'log':
        return '📄';
      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'gif':
      case 'bmp':
        return '🖼️';
      case 'mp3':
      case 'wav':
      case 'ogg':
        return '🎵';
      case 'mp4':
      case 'avi':
      case 'mov':
      case 'mkv':
        return '🎬';
      case 'pdf':
        return '📕';
      case 'doc':
      case 'docx':
        return '📘';
      case 'xls':
      case 'xlsx':
        return '📗';
      case 'ppt':
      case 'pptx':
        return '📙';
      case 'zip':
      case 'tar':
      case 'gz':
      case 'rar':
        return '📦';
      case 'exe':
      case 'dll':
        return '⚙️';
      case 'sh':
      case 'bash':
      case 'bat':
        return '🔧';
      default:
        return '📄';
    }
  };

  // Format file size
  const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    
    return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
  };

  // New functions for drag and drop
  const handleDragStart = (e, file) => {
    setDraggedFile(file);
    
    // Set text data for internal drag and drop
    e.dataTransfer.setData('text/plain', file.name);
    e.dataTransfer.effectAllowed = 'move';
    
    // Create a clone of the file row to use as the drag image
    const fileRow = document.createElement('div');
    fileRow.className = 'drag-file-preview';
    fileRow.innerHTML = `<span>${getFileIcon(file)}</span><span>${file.name}</span>`;
    fileRow.style.position = 'absolute';
    fileRow.style.top = '-1000px';
    document.body.appendChild(fileRow);
    
    // Use the cloned row as drag image
    setTimeout(() => {
      e.dataTransfer.setDragImage(fileRow, 20, 10);
    }, 10);
    
    // Clean up after drag ends
    e.target.addEventListener('dragend', () => {
      if (fileRow.parentNode) {
        document.body.removeChild(fileRow);
      }
    }, { once: true });
  };

  const handleDragOver = (e, file) => {
    e.preventDefault();
    e.stopPropagation();
    
    // Only allow dropping onto directories
    if (file && file.isDirectory) {
      e.dataTransfer.dropEffect = 'move';
      setDropTarget(file);
    } else {
      e.dataTransfer.dropEffect = 'none';
    }
  };

  const handleDragEnter = (e, file) => {
    e.preventDefault();
    e.stopPropagation();
    if (file && file.isDirectory) {
      setDropTarget(file);
    }
  };

  const handleDragLeave = (e) => {
    e.preventDefault();
    e.stopPropagation();
    setDropTarget(null);
  };

  const handleDrop = async (e, targetFile) => {
    e.preventDefault();
    e.stopPropagation();
    setDropTarget(null);

    // If we're dragging from inside the file explorer
    if (draggedFile && targetFile && targetFile.isDirectory) {
      await moveFile(draggedFile, targetFile.path);
      setDraggedFile(null);
      return;
    }

    // If we're dragging from outside the file explorer (file upload)
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      const targetPath = targetFile ? targetFile.path : currentPath;
      await handleFileUpload(e.dataTransfer.files, targetPath);
    }
  };

  const handleFileDrop = async (e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    // Handle files dragged from the user's computer
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      await handleFileUpload(e.dataTransfer.files, currentPath);
    }
  };

  const handleDragEnterContent = (e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  };

  const handleDragLeaveContent = (e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  };

  const moveFile = async (sourceFile, targetDir) => {
    if (!agent || !agent.id) return;
    
    setIsLoading(true);
    try {
      // Use mv command to move the file
      const response = await fetch(`${API_URL}/api/agents/${agent.id}/command`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          command: `mv "${sourceFile.path}" "${targetDir}"`,
          type: 'shell',
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to move file: ${response.status}`);
      }

      const result = await response.json();
      
      if (!result.success) {
        throw new Error(result.error || 'Failed to move file');
      }

      // Reload the current directory to show changes
      loadFileList(currentPath);
    } catch (err) {
      console.error('Error moving file:', err);
      setError(`Failed to move file: ${err.message}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleFileUpload = async (files, targetPath) => {
    if (!agent || !agent.id || files.length === 0) return;
    
    setIsUploading(true);
    setUploadProgress(0);
    
    try {
      // Process each file
      for (let i = 0; i < files.length; i++) {
        const file = files[i];
        const fileReader = new FileReader();
        
        fileReader.onload = async (e) => {
          // Get file content as base64
          const fileContent = e.target.result.split(',')[1];
          
          // Use echo and base64 decode to create the file on the target system
          const response = await fetch(`${API_URL}/api/agents/${agent.id}/command`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              command: `echo "${fileContent}" | base64 -d > "${targetPath}/${file.name}"`,
              type: 'shell',
            }),
          });

          if (!response.ok) {
            throw new Error(`Failed to upload file: ${response.status}`);
          }

          const result = await response.json();
          
          if (!result.success) {
            throw new Error(result.error || 'Failed to upload file');
          }
          
          // Update progress
          setUploadProgress(Math.round(((i + 1) / files.length) * 100));
          
          // If this is the last file, reload the directory
          if (i === files.length - 1) {
            loadFileList(currentPath);
            setIsUploading(false);
          }
        };
        
        fileReader.onerror = () => {
          throw new Error(`Error reading file: ${file.name}`);
        };
        
        // Read the file as data URL (to get base64)
        fileReader.readAsDataURL(file);
      }
    } catch (err) {
      console.error('Error uploading files:', err);
      setError(`Failed to upload files: ${err.message}`);
      setIsUploading(false);
    }
  };

  const handleFileSelection = () => {
    fileInputRef.current.click();
  };

  const handleFileInputChange = (e) => {
    if (e.target.files && e.target.files.length > 0) {
      handleFileUpload(e.target.files, currentPath);
    }
  };

  // Add a direct download button function
  const handleDownloadFile = async (file) => {
    if (!agent || !agent.id || file.isDirectory) return;
    
    try {
      setDownloadingFile(file);
      
      // Execute a base64-encoded cat command to get file contents
      const response = await fetch(`${API_URL}/api/agents/${agent.id}/command`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          command: `cat "${file.path}" | base64`,
          type: 'shell',
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to download file: ${response.status}`);
      }

      const result = await response.json();
      
      if (!result.success) {
        throw new Error(result.error || 'Failed to download file');
      }

      // Create a file object from the base64 data
      let fileContent;
      try {
        // Decode base64 to binary
        fileContent = atob(result.result.trim());
      } catch (e) {
        console.error("Base64 decode error:", e);
        throw new Error("Failed to decode file content");
      }

      // Convert to array buffer
      const arrayBuffer = new ArrayBuffer(fileContent.length);
      const uint8Array = new Uint8Array(arrayBuffer);
      for (let i = 0; i < fileContent.length; i++) {
        uint8Array[i] = fileContent.charCodeAt(i);
      }

      // Determine MIME type for proper file handling
      const mimeType = getMimeType(file.name);

      // Create a Blob with the correct MIME type
      const blob = new Blob([arrayBuffer], { type: mimeType });

      // Create a name for the downloaded file
      const filename = file.name;
      
      // Force download approach that works in most browsers
      try {
        // Modern browsers
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        
        // Important: these attributes help trigger the Save As dialog
        link.download = filename;
        link.href = url;
        link.style.display = 'none';
        
        // Add to DOM to work with Firefox
        document.body.appendChild(link);
        
        // Open in new window/tab to force the Save As dialog on some browsers
        const evt = document.createEvent('MouseEvents');
        evt.initEvent('click', true, true);
        link.dispatchEvent(evt);
        
        // Cleanup
        setTimeout(() => {
          document.body.removeChild(link);
          window.URL.revokeObjectURL(url);
        }, 100);
      } catch (e) {
        console.error('Download error:', e);
        alert('Download failed: ' + e.message);
      }
    } catch (err) {
      console.error('Error downloading file:', err);
      setError(`Failed to download file: ${err.message}`);
    } finally {
      setDownloadingFile(null);
    }
  };

  // Helper function to determine MIME type based on file extension
  const getMimeType = (filename) => {
    const extension = filename.split('.').pop().toLowerCase();
    const mimeTypes = {
      'txt': 'text/plain',
      'html': 'text/html',
      'css': 'text/css',
      'js': 'text/javascript',
      'json': 'application/json',
      'pdf': 'application/pdf',
      'jpg': 'image/jpeg',
      'jpeg': 'image/jpeg',
      'png': 'image/png',
      'gif': 'image/gif',
      'svg': 'image/svg+xml',
      'mp3': 'audio/mpeg',
      'mp4': 'video/mp4',
      'zip': 'application/zip',
      'doc': 'application/msword',
      'docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      'xls': 'application/vnd.ms-excel',
      'xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      'ppt': 'application/vnd.ms-powerpoint',
      'pptx': 'application/vnd.openxmlformats-officedocument.presentationml.presentation'
    };
    
    return mimeTypes[extension] || 'application/octet-stream';
  };

  return (
    <div className="file-explorer">
      <div className="file-explorer-header">
        <h2>File Explorer - {agent.hostname}</h2>
        <button className="close-button" onClick={onClose}>&times;</button>
      </div>
      
      <div className="file-explorer-toolbar">
        <button onClick={goToParentDirectory} disabled={currentPath === '/'}>
          ⬆️ Up
        </button>
        <button onClick={() => loadFileList(currentPath)}>
          🔄 Refresh
        </button>
        <button onClick={handleFileSelection}>
          📤 Upload
        </button>
        <input
          type="file"
          ref={fileInputRef}
          onChange={handleFileInputChange}
          style={{ display: 'none' }}
          multiple
        />
        <div className="breadcrumbs">
          {breadcrumbs.map((crumb, index) => (
            <React.Fragment key={crumb.path}>
              {index > 0 && <span className="breadcrumb-separator">/</span>}
              <span 
                className="breadcrumb"
                onClick={() => handleBreadcrumbClick(crumb)}
              >
                {crumb.name}
              </span>
            </React.Fragment>
          ))}
        </div>
      </div>
      
      <div 
        className={`file-explorer-content ${isDragging ? 'dragging' : ''}`}
        onDragEnter={handleDragEnterContent}
        onDragOver={(e) => e.preventDefault()}
        onDragLeave={handleDragLeaveContent}
        onDrop={handleFileDrop}
      >
        {isUploading && (
          <div className="upload-overlay">
            <div className="upload-progress">
              <div className="upload-progress-bar" style={{ width: `${uploadProgress}%` }}></div>
              <div className="upload-progress-text">Uploading... {uploadProgress}%</div>
            </div>
          </div>
        )}
        
        {isDragging && (
          <div className="drop-overlay">
            <div className="drop-message">
              <span className="drop-icon">📥</span>
              <span>Drop files to upload</span>
            </div>
          </div>
        )}

        {isLoading ? (
          <div className="loading-indicator">Loading files...</div>
        ) : error ? (
          <div className="error-message">{error}</div>
        ) : fileList.length === 0 ? (
          <div className="empty-directory">This directory is empty</div>
        ) : (
          <table className="file-list">
            <thead>
              <tr>
                <th style={{ width: '32px' }}></th>
                <th>Name</th>
                <th>Size</th>
                <th>Type</th>
                <th>Modified</th>
                <th>Permissions</th>
                <th style={{ width: '80px' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {fileList.map((file, index) => (
                <tr 
                  key={index}
                  className={`file-item ${selectedFile === file ? 'selected' : ''} ${file.isDirectory ? 'directory' : 'file'} ${dropTarget === file ? 'drop-target' : ''}`}
                  onClick={() => handleFileClick(file)}
                  onDoubleClick={() => file.isDirectory && handleFileClick(file)}
                  draggable
                  onDragStart={(e) => handleDragStart(e, file)}
                  onDragOver={(e) => handleDragOver(e, file)}
                  onDragEnter={(e) => handleDragEnter(e, file)}
                  onDragLeave={handleDragLeave}
                  onDrop={(e) => handleDrop(e, file)}
                >
                  <td className="file-icon">{getFileIcon(file)}</td>
                  <td className="file-name">{file.name}</td>
                  <td className="file-size">{file.isDirectory ? '--' : formatFileSize(file.size)}</td>
                  <td className="file-type">{file.isDirectory ? 'Directory' : file.isLink ? 'Link' : 'File'}</td>
                  <td className="file-modified">{file.modifiedTime}</td>
                  <td className="file-permissions">{file.permissions}</td>
                  <td className="file-actions">
                    {!file.isDirectory && (
                      <button 
                        className={`action-button download ${downloadingFile === file ? 'downloading' : ''}`}
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDownloadFile(file);
                        }}
                        title="Download file"
                        disabled={downloadingFile === file}
                      >
                        {downloadingFile === file ? 'Downloading...' : 'Download ⬇️'}
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
      
      <div className="file-explorer-statusbar">
        <div>{fileList.length} items • Current path: {currentPath}</div>
        <div className="statusbar-info">
          <span className="drag-drop-info">✨ Tip: Drag files between folders to move them. Click Download to save files to your computer.</span>
        </div>
      </div>
    </div>
  );
};

export default FileExplorer; 