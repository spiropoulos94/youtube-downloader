import React, { useState, useEffect } from "react";
import {
  Alert,
  Box,
  Button,
  Container,
  Paper,
  Snackbar,
  TextField,
  Typography,
  Stack,
  ThemeProvider,
  createTheme,
  Grid,
  IconButton,
  Tooltip,
} from "@mui/material";
import { DownloadableVideo, TaskStatus } from "./types";
import Downloadable from "./components/Downloadable";
import { downloadVideo } from "./utils/api";
import DeleteIcon from "@mui/icons-material/Delete";
import YouTubeIcon from "@mui/icons-material/YouTube";
import ErrorIcon from "@mui/icons-material/Error";

// Create a custom theme
const theme = createTheme({
  palette: {
    primary: {
      main: "#FF0000", // YouTube red
    },
    secondary: {
      main: "#282828", // Dark gray
    },
    background: {
      default: "#f9f9f9",
      paper: "#ffffff",
    },
    success: {
      main: "#2196f3", // Using blue instead of green
      light: "#e3f2fd",
    },
    error: {
      main: "#f44336",
      light: "#ffebee",
    },
  },
  typography: {
    fontFamily: "'Roboto', 'Helvetica', 'Arial', sans-serif",
    h4: {
      fontWeight: 700,
    },
    h5: {
      fontWeight: 600,
    },
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          textTransform: "none",
          fontWeight: 600,
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          borderRadius: 12,
        },
      },
    },
  },
});

// Add a helper function to format error messages
const formatErrorMessage = (
  error: string
): { title: string; description: string } => {
  // Common error messages and their user-friendly versions
  if (error === "not a YouTube URL") {
    return {
      title: "Invalid YouTube URL",
      description:
        "Please enter a valid YouTube video URL (e.g., https://www.youtube.com/watch?v=...)",
    };
  }
  if (error.includes("Failed to download video")) {
    return {
      title: "Download Failed",
      description:
        "Unable to download the video. Please check if the video is available and try again.",
    };
  }
  if (error === "URL cannot be empty") {
    return {
      title: "Missing URL",
      description: "Please enter a YouTube video URL to download.",
    };
  }
  // Default error message
  return {
    title: "Error",
    description: error,
  };
};

const App: React.FC = () => {
  // Function to ensure video statuses are valid enum values
  const ensureValidStatuses = (
    savedVideos: DownloadableVideo[]
  ): DownloadableVideo[] => {
    return savedVideos.map((video) => ({
      ...video,
      status: video.status as TaskStatus,
    }));
  };

  const [url, setUrl] = useState<string>("");
  const [videos, setVideos] = useState<DownloadableVideo[]>(() => {
    const savedVideos = localStorage.getItem("youtube-downloads");
    return savedVideos ? ensureValidStatuses(JSON.parse(savedVideos)) : [];
  });
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Save videos to localStorage whenever they change
  useEffect(() => {
    localStorage.setItem("youtube-downloads", JSON.stringify(videos));
  }, [videos]);

  const handleUrlChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUrl(e.target.value);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!url.trim()) {
      setError("URL cannot be empty");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await downloadVideo(url);
      const newVideo: DownloadableVideo = {
        taskId: response.data.task_id,
        url,
        status: TaskStatus.TaskStatusPending,
      };
      setVideos((prev) => [newVideo, ...prev]);
      setUrl("");
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to download video");
      console.error("Download error:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleStatusUpdate = (
    taskId: string,
    status: TaskStatus,
    error?: string
  ) => {
    setVideos((prev) =>
      prev.map((video) =>
        video.taskId === taskId ? { ...video, status, error } : video
      )
    );
  };

  const handleDeleteAll = () => {
    setVideos([]);
  };

  const handleDeleteVideo = (taskId: string) => {
    setVideos((prev) => prev.filter((video) => video.taskId !== taskId));
  };

  return (
    <ThemeProvider theme={theme}>
      <Box sx={{ minHeight: "100vh", bgcolor: "background.default", py: 4 }}>
        <Container maxWidth="lg">
          <Paper
            elevation={3}
            sx={{
              p: 4,
              mb: 4,
              background: "linear-gradient(45deg, #FF0000 30%, #FF5252 90%)",
              color: "white",
            }}
          >
            <Stack spacing={3}>
              <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
                <YouTubeIcon sx={{ fontSize: 40 }} />
                <Typography variant="h4" component="h1">
                  YouTube Video Downloader
                </Typography>
              </Box>

              <Box component="form" onSubmit={handleSubmit}>
                <Stack spacing={2}>
                  <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
                    <TextField
                      fullWidth
                      variant="outlined"
                      value={url}
                      onChange={handleUrlChange}
                      placeholder="https://www.youtube.com/watch?v=..."
                      disabled={loading}
                      error={!!error}
                      sx={{
                        bgcolor: "white",
                        borderRadius: 2,
                        "& .MuiOutlinedInput-root": {
                          borderRadius: 2,
                        },
                      }}
                    />
                    <Button
                      variant="contained"
                      color="secondary"
                      type="submit"
                      disabled={loading || !url.trim()}
                      sx={{
                        minWidth: { xs: "100%", sm: "200px" },
                        py: { xs: 1.5, sm: "auto" },
                      }}
                    >
                      {loading ? "Processing..." : "Download"}
                    </Button>
                  </Stack>
                  {error && (
                    <Paper
                      elevation={0}
                      sx={{
                        p: 2,
                        bgcolor: "error.light",
                        borderRadius: 2,
                        border: 1,
                        borderColor: "error.main",
                      }}
                    >
                      <Stack spacing={0.5}>
                        <Typography
                          variant="subtitle2"
                          color="error"
                          sx={{
                            display: "flex",
                            alignItems: "center",
                            gap: 1,
                            fontWeight: 600,
                          }}
                        >
                          <ErrorIcon fontSize="small" />
                          {formatErrorMessage(error).title}
                        </Typography>
                        <Typography
                          variant="body2"
                          color="error.dark"
                          sx={{ opacity: 0.9 }}
                        >
                          {formatErrorMessage(error).description}
                        </Typography>
                      </Stack>
                    </Paper>
                  )}
                </Stack>
              </Box>
            </Stack>
          </Paper>

          {videos.length > 0 && (
            <Box>
              <Box
                sx={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  mb: 3,
                }}
              >
                <Typography variant="h5" component="h2">
                  Your Downloads ({videos.length})
                </Typography>
                <Tooltip title="Delete all downloads">
                  <IconButton
                    onClick={handleDeleteAll}
                    color="error"
                    sx={{ "&:hover": { transform: "scale(1.1)" } }}
                  >
                    <DeleteIcon />
                  </IconButton>
                </Tooltip>
              </Box>

              <Grid container spacing={3}>
                {videos.map((video) => (
                  <Grid item xs={12} sm={6} md={4} key={video.taskId}>
                    <Downloadable
                      video={video}
                      onStatusUpdate={handleStatusUpdate}
                      onDelete={() => handleDeleteVideo(video.taskId)}
                    />
                  </Grid>
                ))}
              </Grid>
            </Box>
          )}

          <Snackbar
            open={!!error}
            autoHideDuration={6000}
            onClose={() => setError(null)}
            anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
          >
            <Alert severity="error" onClose={() => setError(null)}>
              {error}
            </Alert>
          </Snackbar>
        </Container>
      </Box>
    </ThemeProvider>
  );
};

export default App;
