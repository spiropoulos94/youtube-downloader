export enum TaskStatus {
  TaskStatusPending = "pending",
  TaskStatusInProgress = "in_progress",
  TaskStatusCompleted = "completed",
  TaskStatusFailed = "failed",
}

export interface DownloadRequest {
  url: string;
}

export interface DownloadResponse {
  success: boolean;
  data: {
    task_id: string;
  };
}

export interface TaskStatusResponseData {
  status: TaskStatus;
  file_path?: string;
  download_url?: string;
  error?: string;
  title?: string;
  thumbnail_url?: string;
  duration?: string;
}

export interface TaskStatusResponse {
  success: boolean;
  data: TaskStatusResponseData;
}

export interface DownloadableVideo {
  taskId: string;
  url: string;
  status: TaskStatus;
  error?: string;
  title?: string;
  thumbnailUrl?: string;
  duration?: string;
  downloadUrl?: string;
}
