import axios from "axios";
import {
  DownloadRequest,
  DownloadResponse,
  TaskStatusResponse,
  TaskStatusResponseData,
} from "../types";

const API_URL = "/api";

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
  return response.data.data;
};

export const getVideoDownloadUrl = (taskId: string): string => {
  return `${API_URL}/videos/${taskId}`;
};
