import axios from "axios";
import { downloadVideo, getTaskStatus, getVideoDownloadUrl } from "../api";
import { TaskStatus } from "../../types";

// Mock axios
jest.mock("axios");
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe("API utilities", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe("downloadVideo", () => {
    it("should send a POST request to download a video and return the response data", async () => {
      // Mock response data
      const mockResponseData = {
        success: true,
        data: {
          task_id: "test-task-123",
        },
      };

      // Setup axios mock
      mockedAxios.post.mockResolvedValueOnce({ data: mockResponseData });

      // Call the function
      const result = await downloadVideo(
        "https://www.youtube.com/watch?v=test123"
      );

      // Assertions
      expect(mockedAxios.post).toHaveBeenCalledWith("/api/download", {
        url: "https://www.youtube.com/watch?v=test123",
      });
      expect(result).toEqual(mockResponseData);
    });

    it("should throw an error when the API call fails", async () => {
      // Setup axios mock to reject
      const mockError = new Error("Network error");
      mockedAxios.post.mockRejectedValueOnce(mockError);

      // Call the function and expect it to throw
      await expect(
        downloadVideo("https://www.youtube.com/watch?v=test123")
      ).rejects.toThrow("Network error");

      // Verify the POST was attempted
      expect(mockedAxios.post).toHaveBeenCalledWith("/api/download", {
        url: "https://www.youtube.com/watch?v=test123",
      });
    });
  });

  describe("getTaskStatus", () => {
    it("should send a GET request to get task status and return the response data", async () => {
      // Mock response data
      const mockResponseData = {
        success: true,
        data: {
          status: TaskStatus.TaskStatusCompleted,
          title: "Test Video",
          thumbnail_url: "https://example.com/thumbnail.jpg",
          duration: "10:30",
          download_url: "/api/videos/test-task-123",
        },
      };

      // Setup axios mock
      mockedAxios.get.mockResolvedValueOnce({ data: mockResponseData });

      // Call the function
      const result = await getTaskStatus("test-task-123");

      // Assertions
      expect(mockedAxios.get).toHaveBeenCalledWith("/api/tasks/test-task-123");
      expect(result).toEqual(mockResponseData.data);
    });

    it("should throw an error when the API call fails", async () => {
      // Setup axios mock to reject
      const mockError = new Error("Network error");
      mockedAxios.get.mockRejectedValueOnce(mockError);

      // Call the function and expect it to throw
      await expect(getTaskStatus("test-task-123")).rejects.toThrow(
        "Network error"
      );

      // Verify the GET was attempted
      expect(mockedAxios.get).toHaveBeenCalledWith("/api/tasks/test-task-123");
    });
  });

  describe("getVideoDownloadUrl", () => {
    it("should return the cached download URL if available", async () => {
      // Setup cache by calling getTaskStatus first
      const mockResponseData = {
        success: true,
        data: {
          status: TaskStatus.TaskStatusCompleted,
          download_url: "/custom/download/url",
        },
      };

      // Manually populate the cache by calling getTaskStatus
      mockedAxios.get.mockResolvedValueOnce({ data: mockResponseData });
      await getTaskStatus("test-task-123");

      // Now get the download URL
      const url = getVideoDownloadUrl("test-task-123");

      // Assertions
      expect(url).toBe("/custom/download/url");
    });

    it("should return the fallback URL when cache is not available", () => {
      // Call the function without populating cache
      const url = getVideoDownloadUrl("test-task-456");

      // Assertions - using the full URL with domain
      expect(url).toBe("http://localhost:8080/api/videos/test-task-456");
    });
  });
});
