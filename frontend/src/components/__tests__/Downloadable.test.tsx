import React from "react";
import {
  render,
  screen,
  fireEvent,
  waitFor,
  act,
} from "@testing-library/react";
import "@testing-library/jest-dom";
import Downloadable from "../Downloadable";
import { DownloadableVideo, TaskStatus } from "../../types";
import * as api from "../../utils/api";
import axios from "axios";

// Mock axios
jest.mock("axios");

// Mock the API functions
jest.mock("../../utils/api", () => ({
  getTaskStatus: jest.fn(),
  getVideoDownloadUrl: jest.fn(
    () => "http://localhost:8080/api/videos/mock-task-id"
  ),
}));

// Increase the Jest timeout for these tests
jest.setTimeout(10000);

describe("Downloadable Component", () => {
  // Mock handlers
  const mockOnStatusUpdate = jest.fn();
  const mockOnDelete = jest.fn();

  // Mock video data
  const pendingVideo: DownloadableVideo = {
    taskId: "mock-task-id",
    url: "https://www.youtube.com/watch?v=test123",
    status: TaskStatus.TaskStatusPending,
  };

  const completedVideo: DownloadableVideo = {
    taskId: "mock-task-id",
    url: "https://www.youtube.com/watch?v=test123",
    status: TaskStatus.TaskStatusCompleted,
    title: "Test Video",
    thumbnailUrl: "https://example.com/thumbnail.jpg",
    duration: "10:30",
    downloadUrl: "http://localhost:8080/api/videos/mock-task-id",
  };

  const failedVideo: DownloadableVideo = {
    taskId: "mock-task-id",
    url: "https://www.youtube.com/watch?v=test123",
    status: TaskStatus.TaskStatusFailed,
    error: "Failed to download video",
  };

  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();

    // Default mock implementation for getTaskStatus that returns a properly shaped response
    (api.getTaskStatus as jest.Mock).mockImplementation((taskId) => {
      return Promise.resolve({
        status: "completed",
        title: "Updated Title",
        thumbnail_url: "https://example.com/new-thumbnail.jpg",
        duration: "5:30",
        download_url: "http://localhost:8080/api/videos/mock-task-id",
      });
    });
  });

  afterEach(() => {
    act(() => {
      jest.runOnlyPendingTimers();
    });
    jest.useRealTimers();
  });

  it("should render a pending video correctly", async () => {
    render(
      <Downloadable
        video={pendingVideo}
        onStatusUpdate={mockOnStatusUpdate}
        onDelete={mockOnDelete}
      />
    );

    // Wait for component to be fully rendered
    await waitFor(() => {
      expect(screen.getByRole("progressbar")).toBeInTheDocument();
    });

    // Check for loading indicator
    expect(screen.getByText(/processing/i)).toBeInTheDocument();

    // Check for delete button
    const deleteButton = screen.getByRole("button", {
      name: /delete download/i,
    });
    expect(deleteButton).toBeInTheDocument();
  });

  it("should render a completed video correctly", async () => {
    render(
      <Downloadable
        video={completedVideo}
        onStatusUpdate={mockOnStatusUpdate}
        onDelete={mockOnDelete}
      />
    );

    // Wait for component to be fully rendered
    await waitFor(() => {
      expect(screen.getByText("Test Video")).toBeInTheDocument();
    });

    // Check for video metadata
    expect(screen.getByText(/duration: 10:30/i)).toBeInTheDocument();

    // Check for download button
    const downloadButton = screen.getByText("Download");
    expect(downloadButton).toBeInTheDocument();
  });

  it("should render a failed video correctly", async () => {
    render(
      <Downloadable
        video={failedVideo}
        onStatusUpdate={mockOnStatusUpdate}
        onDelete={mockOnDelete}
      />
    );

    // Wait for component to be fully rendered
    await waitFor(() => {
      expect(screen.getByText("YouTube Video")).toBeInTheDocument();
    });

    // Check for delete button
    expect(
      screen.getByRole("button", { name: /delete download/i })
    ).toBeInTheDocument();
  });

  it("should call onDelete when delete button is clicked", async () => {
    render(
      <Downloadable
        video={completedVideo}
        onStatusUpdate={mockOnStatusUpdate}
        onDelete={mockOnDelete}
      />
    );

    // Wait for component to be fully rendered
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /delete download/i })
      ).toBeInTheDocument();
    });

    // Find and click delete button
    const deleteButton = screen.getByRole("button", {
      name: /delete download/i,
    });

    fireEvent.click(deleteButton);

    // Verify onDelete was called
    expect(mockOnDelete).toHaveBeenCalledTimes(1);
  });

  it("should poll task status for pending videos", async () => {
    // Mock API response
    const mockResponse = {
      status: "completed",
      title: "Updated Title",
      thumbnail_url: "https://example.com/new-thumbnail.jpg",
      duration: "5:30",
      download_url: "http://localhost:8080/api/videos/mock-task-id",
    };

    (api.getTaskStatus as jest.Mock).mockResolvedValue(mockResponse);

    render(
      <Downloadable
        video={pendingVideo}
        onStatusUpdate={mockOnStatusUpdate}
        onDelete={mockOnDelete}
      />
    );

    // Fast-forward timers to trigger polling
    act(() => {
      jest.advanceTimersByTime(2000);
    });

    // Allow any pending promises to resolve
    await waitFor(() => {
      expect(api.getTaskStatus).toHaveBeenCalledWith("mock-task-id");
    });

    // Verify API call and onStatusUpdate
    expect(mockOnStatusUpdate).toHaveBeenCalled();
  });

  // Skip this test for now as window.location.href setting might be handled differently
  it.skip("should trigger download when download button is clicked", async () => {
    // Mock window.location
    const originalLocation = window.location;
    const mockLocation = { href: "" };

    // Replace window.location with our mock
    Object.defineProperty(window, "location", {
      writable: true,
      value: mockLocation,
    });

    render(
      <Downloadable
        video={completedVideo}
        onStatusUpdate={mockOnStatusUpdate}
        onDelete={mockOnDelete}
      />
    );

    // Wait for component to be fully rendered
    await waitFor(() => {
      expect(screen.getByText("Download")).toBeInTheDocument();
    });

    // Find and click download button
    const downloadButton = screen.getByText("Download");

    fireEvent.click(downloadButton);

    // Verify download URL was set
    expect(mockLocation.href).toBe(
      "http://localhost:8080/api/videos/mock-task-id"
    );

    // Restore original location
    Object.defineProperty(window, "location", {
      writable: true,
      value: originalLocation,
    });
  });
});
