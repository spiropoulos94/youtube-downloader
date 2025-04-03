package services

type Services struct {
	YouTube  YouTubeServiceInterface
	Cleanup  CleanupServiceInterface
	Frontend FrontendServiceInterface
}
