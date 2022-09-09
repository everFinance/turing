package tracker

// func TestTracker(t *testing.T) {
// 	tracker := New([]types.Tag{
// 		types.Tag{Name: "App", Value: "everToken"},
// 		types.Tag{Name: "Owner", Value: "dQzTM9hXV5MD1fRniOKI3MvPF_-8b2XDLmpfcMN9hi8"},
// 	}, DefaultNodeUrl)
// 	tracker.Run()
// 	fmt.Println("123")

// 	go func() {
// 		for {
// 			tx := <-tracker.SubscribeTx()
// 			fmt.Printf("from stream: %+v\n%+v\n%+v\n\n", tx.ID, tx.Owner, string(tx.Data))
// 		}
// 	}()

// 	for {
// 		time.Sleep(1 * time.Second)
// 		if tracker.IsSynced {
// 			tracker.Close()
// 			return
// 		}
// 	}
// }
