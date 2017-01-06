package main

type Watch struct {
	ProjectID string
	BucketName string
	TopicName string
}

func NewWatch() Watch {
	return Watch{
		ProjectID:  os.Getenv("PROJECT"),
		BucketName: os.Getenv("BUCKET"),
		TopicName: os.Getenv("TOPIC"),
	}
}
