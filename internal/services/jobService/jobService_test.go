package jobservice

// func TestWorkerProcessesJobFromQueue(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	st := storage.NewInMemoryJobStorage()
// 	mockProc := mocks.NewMockProcessor(ctrl)

// 	called := make(chan struct{}, 1)
// 	mockProc.EXPECT().Process(gomock.Any(), gomock.Any()).DoAndReturn(
// 		func(ctx context.Context, req models.AnalyzeRequest) (models.SummarizeResponse, error) {
// 			called <- struct{}{}
// 			return models.SummarizeResponse{ResponseMessage: "ok"}, nil
// 		},
// 	).Times(1)

// 	svc := NewJobQueueService(10, 1, st, mockProc)

// 	job := models.Job{ID: "job-test-1"}
// 	select {
// 	case svc.queue <- job:
// 	default:
// 		t.Fatal("failed to enqueue job into internal queue")
// 	}

// 	require.Eventually(t, func() bool {
// 		select {
// 		case <-called:
// 			return true
// 		default:
// 			return false
// 		}
// 	}, 2*time.Second, 10*time.Millisecond)
// }

// func TestMultipleJobsProcessedConcurrently(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	st := storage.NewInMemoryJobStorage()
// 	mockProc := mocks.NewMockProcessor(ctrl)

// 	const n = 10
// 	called := make(chan struct{}, n)
// 	mockProc.EXPECT().Process(gomock.Any(), gomock.Any()).DoAndReturn(
// 		func(ctx context.Context, req models.AnalyzeRequest) (models.SummarizeResponse, error) {
// 			called <- struct{}{}
// 			return models.SummarizeResponse{ResponseMessage: "ok"}, nil
// 		},
// 	).Times(n)

// 	svc := NewJobQueueService(50, 4, st, mockProc)

// 	for i := 0; i < n; i++ {
// 		j := models.Job{ID: "job-" + string(rune(i+65))}
// 		select {
// 		case svc.queue <- j:
// 		default:
// 			t.Fatalf("failed to enqueue job %d", i)
// 		}
// 	}

// 	// wait until all calls recorded
// 	require.Eventually(t, func() bool { return len(called) == n }, 3*time.Second, 20*time.Millisecond)
// }
