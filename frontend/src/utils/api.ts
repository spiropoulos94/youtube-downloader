import axios from "axios";
import {
  DownloadRequest,
  DownloadResponse,
  TaskStatusResponse,
  TaskStatusResponseData,
} from "../types";

const API_URL = "/api";
const BACKEND_URL = "http://localhost:8080";

export const downloadVideo = async (url: string): Promise<DownloadResponse> => {
  const response = await axios.post<DownloadResponse>(`${API_URL}/download`, {
    url,
  } as DownloadRequest);
  return response.data;
};

export const getTaskStatus = async (
  taskId: string
): Promise<TaskStatusResponseData> => {
  console.log("Getting task status for:", taskId);
  console.log("API URL:", `${API_URL}/tasks/${taskId}`);
  const response = await axios.get<TaskStatusResponse>(
    `${API_URL}/tasks/${taskId}`
  );
  console.log("Task status response:", response.data, taskId);
  return response.data.data;
};

export const getVideoDownloadUrl = (taskId: string): string => {
  return `${BACKEND_URL}/api/videos/${taskId}`;
};
