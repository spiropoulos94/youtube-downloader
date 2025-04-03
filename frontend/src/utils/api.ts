import axios from "axios";
import {
  DownloadRequest,
  DownloadResponse,
  TaskStatusResponse,
  TaskStatusResponseData,
} from "../types";

const API_URL = "/api";
// Use the backend URL from environment variables or fallback to localhost
const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

// Store task status responses for caching download URLs
const taskStatusCache: Record<string, TaskStatusResponseData> = {};

export const downloadVideo = async (url: string): Promise<DownloadResponse> => {
  const response = await axios.post<DownloadResponse>(`${API_URL}/download`, {
    url,
  } as DownloadRequest);
  return response.data;
};

export const getTaskStatus = async (
  taskId: string
): Promise<TaskStatusResponseData> => {
  const response = await axios.get<TaskStatusResponse>(
    `${API_URL}/tasks/${taskId}`
  );
  // Cache the response data
  taskStatusCache[taskId] = response.data.data;
  return response.data.data;
};

export const getVideoDownloadUrl = (taskId: string): string => {
  // Check if we have a cached response with a download_url
  if (taskStatusCache[taskId]?.download_url) {
    return taskStatusCache[taskId].download_url!;
  }

  // Fallback to the original URL construction
  return `${BACKEND_URL}${API_URL}/videos/${taskId}`;
};
