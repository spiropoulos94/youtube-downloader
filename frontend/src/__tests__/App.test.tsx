import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import "@testing-library/jest-dom";
import userEvent from "@testing-library/user-event";
import App from "../App";
import {
  downloadVideo,
  getTaskStatus,
  getVideoDownloadUrl,
} from "../utils/api";
import { TaskStatus } from "../types";

// Mock the API functions
jest.mock("../utils/api", () => ({
  downloadVideo: jest.fn(),
  getTaskStatus: jest.fn(),
  getVideoDownloadUrl: jest.fn(),
}));

// Mock localStorage
const mockLocalStorage = (function () {
  let store: Record<string, string> = {};
  return {
    getItem: function (key: string) {
      return store[key] || null;
    },
    setItem: function (key: string, value: string) {
      store[key] = value;
    },
    clear: function () {
      store = {};
    },
  };
})();

Object.defineProperty(window, "localStorage", {
  value: mockLocalStorage,
});

describe("App Component", () => {
  beforeEach(() => {
    // Clear all mocks before each test
    jest.clearAllMocks();
    mockLocalStorage.clear();

    // Default mock implementations
    (downloadVideo as jest.Mock).mockResolvedValue({
      data: {
        task_id: "test-task-id",
        message: "Download started",
      },
    });

    (getTaskStatus as jest.Mock).mockResolvedValue({
      status: TaskStatus.TaskStatusPending,
      title: "Test Video",
      thumbnail: "test-thumbnail.jpg",
      url: "https://www.youtube.com/watch?v=test",
      taskId: "test-task-id",
    });

    (getVideoDownloadUrl as jest.Mock).mockResolvedValue(
      "http://example.com/download/test-video.mp4"
    );
  });

  test("renders application header", async () => {
    render(<App />);

    expect(screen.getByText("YouTube Video Downloader")).toBeInTheDocument();
    const input = screen.getByPlaceholderText(
      "https://www.youtube.com/watch?v=..."
    );
    expect(input).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /download/i })
    ).toBeInTheDocument();
  });

  test("handles input change", async () => {
    render(<App />);

    const input = screen.getByPlaceholderText(
      "https://www.youtube.com/watch?v=..."
    );
    await userEvent.type(input, "https://www.youtube.com/watch?v=test");

    expect(input).toHaveValue("https://www.youtube.com/watch?v=test");
  });

  test("handles form submission", async () => {
    render(<App />);

    const input = screen.getByPlaceholderText(
      "https://www.youtube.com/watch?v=..."
    );
    await userEvent.type(input, "https://www.youtube.com/watch?v=test");

    const submitButton = screen.getByRole("button", { name: /download/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(downloadVideo).toHaveBeenCalledWith(
        "https://www.youtube.com/watch?v=test"
      );
    });

    // Wait for the form submission to complete
    await waitFor(() => {
      expect(input).toHaveValue("");
    });
  });

  test("displays error message for API errors", async () => {
    // Mock API error
    (downloadVideo as jest.Mock).mockRejectedValue({
      response: {
        data: {
          error: "Invalid YouTube URL",
        },
      },
    });

    render(<App />);

    const input = screen.getByPlaceholderText(
      "https://www.youtube.com/watch?v=..."
    );
    await userEvent.type(input, "https://www.youtube.com/watch?v=test");

    const submitButton = screen.getByRole("button", { name: /download/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Invalid YouTube URL")).toBeInTheDocument();
    });
  });

  test("restores downloads from localStorage", async () => {
    // Set up localStorage with saved downloads
    const savedDownloads = JSON.stringify([
      {
        taskId: "saved-task-id",
        status: TaskStatus.TaskStatusCompleted,
        title: "Saved Video",
        thumbnail: "saved-thumbnail.jpg",
        url: "https://www.youtube.com/watch?v=saved",
      },
    ]);

    mockLocalStorage.setItem("youtube-downloads", savedDownloads);

    render(<App />);

    // In this case we need to wait for the component to load the data from localStorage
    await waitFor(() => {
      expect(screen.getByText("Saved Video")).toBeInTheDocument();
    });
  });

  test("does not submit form with empty URL", async () => {
    render(<App />);

    // Try to submit with empty input
    const submitButton = screen.getByRole("button", { name: /download/i });

    // The button should be disabled due to empty input
    expect(submitButton).toBeDisabled();

    // Check that the download function is not called
    expect(downloadVideo).not.toHaveBeenCalled();
  });
});
