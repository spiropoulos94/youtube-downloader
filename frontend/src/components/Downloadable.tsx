import React, { useEffect, useState } from "react";
import {
  Box,
  Button,
  Card,
  CardActions,
  CardContent,
  CircularProgress,
  Typography,
  IconButton,
  styled,
} from "@mui/material";
import { DownloadableVideo, TaskStatus } from "../types";
import { getTaskStatus, getVideoDownloadUrl } from "../utils/api";
import FileDownloadIcon from "@mui/icons-material/FileDownload";
import ErrorIcon from "@mui/icons-material/Error";
import DeleteIcon from "@mui/icons-material/Delete";
import YouTubeIcon from "@mui/icons-material/YouTube";

interface DownloadableProps {
  video: DownloadableVideo;
  onStatusUpdate: (
    taskId: string,
    status: TaskStatus,
    error?: string,
    title?: string,
    thumbnailUrl?: string,
    duration?: string
  ) => void;
  onDelete: () => void;
}

const StyledCard = styled(Card)(({ theme }) => ({
  height: "100%",
  display: "flex",
  flexDirection: "column",
  transition: "transform 0.2s ease-in-out",
  "&:hover": {
    transform: "translateY(-4px)",
  },
  position: "relative",
  minHeight: 280,
  borderRadius: 16,
  backgroundColor: "#fff5f5", // Very light red background
}));

const DeleteButton = styled(IconButton)(({ theme }) => ({
  position: "absolute",
  top: 8,
  right: 8,
  zIndex: 10,
  backgroundColor: "rgba(255, 255, 255, 0.8)",
  "&:hover": {
    backgroundColor: theme.palette.error.light,
    transform: "scale(1.1)",
  },
  transition: "transform 0.2s ease-in-out, background-color 0.2s ease-in-out",
}));

const Downloadable: React.FC<DownloadableProps> = ({
  video,
  onStatusUpdate,
  onDelete,
}) => {
  const [loading, setLoading] = useState<boolean>(true);
  const [pollingInterval, setPollingInterval] = useState<NodeJS.Timeout | null>(
    null
  );

  const isCompleted = video.status === TaskStatus.TaskStatusCompleted;
  const isFailed = video.status === TaskStatus.TaskStatusFailed;
  const isInProgress =
    video.status === TaskStatus.TaskStatusPending ||
    video.status === TaskStatus.TaskStatusInProgress;

  const getVideoId = (url: string): string => {
    try {
      return url.split("v=")[1].split("&")[0] || "unknown";
    } catch {
      return "unknown";
    }
  };

  const pollTaskStatus = async () => {
    try {
      const response = await getTaskStatus(video.taskId);
      const statusValue = response.status;
      let status: TaskStatus;

      if (statusValue === "completed") {
        status = TaskStatus.TaskStatusCompleted;
      } else if (statusValue === "failed") {
        status = TaskStatus.TaskStatusFailed;
      } else if (statusValue === "in_progress") {
        status = TaskStatus.TaskStatusInProgress;
      } else {
        status = TaskStatus.TaskStatusPending;
      }

      // Update the local state with metadata
      const updatedVideo = {
        ...video,
        status,
        error: response.error,
        title: response.title,
        thumbnailUrl: response.thumbnail_url,
        duration: response.duration,
      };

      if (
        status !== video.status ||
        updatedVideo.title !== video.title ||
        updatedVideo.thumbnailUrl !== video.thumbnailUrl
      ) {
        onStatusUpdate(
          video.taskId,
          status,
          response.error,
          response.title,
          response.thumbnail_url,
          response.duration
        );
      }

      if (
        status === TaskStatus.TaskStatusCompleted ||
        status === TaskStatus.TaskStatusFailed
      ) {
        if (pollingInterval) {
          clearInterval(pollingInterval);
          setPollingInterval(null);
        }
        setLoading(false);
      }
    } catch (error) {
      console.error("Error polling task status:", error);
    }
  };

  useEffect(() => {
    pollTaskStatus();

    if (
      video.status === TaskStatus.TaskStatusPending ||
      video.status === TaskStatus.TaskStatusInProgress
    ) {
      const interval = setInterval(pollTaskStatus, 2000);
      setPollingInterval(interval);

      return () => {
        if (interval) clearInterval(interval);
      };
    } else {
      setLoading(false);
    }
  }, [video.taskId, video.status]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleDownload = () => {
    window.location.href = getVideoDownloadUrl(video.taskId);
  };

  return (
    <StyledCard
      sx={{
        bgcolor: isFailed
          ? "error.light"
          : isCompleted
          ? "#fff5f5" // Very light red for completed state
          : "#fff5f5", // Very light red for default state
        borderColor: isFailed
          ? "error.main"
          : isCompleted
          ? "primary.main"
          : "divider",
        borderWidth: 1,
        borderStyle: "solid",
      }}
    >
      <DeleteButton
        onClick={onDelete}
        size="small"
        color="error"
        aria-label="delete download"
      >
        <DeleteIcon />
      </DeleteButton>

      <CardContent
        sx={{
          flexGrow: 1,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          position: "relative",
        }}
      >
        {isInProgress ? (
          <Box sx={{ textAlign: "center" }}>
            <CircularProgress
              size={60}
              sx={{
                mb: 2,
                color: "primary.main",
              }}
            />
            <Typography
              variant="body1"
              color="text.primary"
              sx={{ fontWeight: 500 }}
            >
              Processing...
            </Typography>
          </Box>
        ) : (
          <>
            {video.thumbnailUrl ? (
              <Box
                component="img"
                src={video.thumbnailUrl}
                alt={video.title || "YouTube Video"}
                sx={{
                  width: "100%",
                  maxHeight: 160,
                  objectFit: "cover",
                  borderRadius: 1,
                  mb: 2,
                }}
              />
            ) : (
              <YouTubeIcon
                sx={{
                  fontSize: 60,
                  mb: 2,
                  color: "primary.main",
                }}
              />
            )}
            <Typography
              variant="h6"
              component="div"
              align="center"
              gutterBottom
              sx={{
                color: "text.primary",
                fontWeight: 600,
                // Ensure long titles don't overflow
                overflow: "hidden",
                textOverflow: "ellipsis",
                display: "-webkit-box",
                WebkitLineClamp: 2,
                WebkitBoxOrient: "vertical",
              }}
            >
              {video.title || getVideoId(video.url)}
            </Typography>
            {video.duration && (
              <Typography
                variant="body2"
                color="text.secondary"
                sx={{ mt: -1, mb: 1 }}
              >
                Duration: {video.duration}
              </Typography>
            )}
          </>
        )}

        {isFailed && (
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              mt: 2,
              color: "error.main",
              bgcolor: "error.light",
              p: 1,
              borderRadius: 1,
              width: "100%",
              justifyContent: "center",
            }}
          >
            <ErrorIcon sx={{ mr: 1 }} />
            <Typography variant="body2" color="error" sx={{ fontWeight: 500 }}>
              {video.error || "Download failed"}
            </Typography>
          </Box>
        )}
      </CardContent>

      {isCompleted && (
        <CardActions sx={{ p: 2, pt: 0 }}>
          <Button
            variant="contained"
            color="primary"
            startIcon={<FileDownloadIcon />}
            onClick={handleDownload}
            fullWidth
            sx={{
              borderRadius: 8,
              "&:hover": {
                bgcolor: "primary.dark",
              },
            }}
          >
            Download
          </Button>
        </CardActions>
      )}
    </StyledCard>
  );
};

export default Downloadable;
