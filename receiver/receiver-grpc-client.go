package receiver

//
// func (r *Receiver) connectToClassifier() (proto.EventServiceClient, error) {
// 	classifierConn, err := grpc.Dial(
// 		r.config.App.Receiver.Classifier.Address,
// 		grpc.WithInsecure(),
// 		grpc.WithStatsHandler(&grpc_server.ConnStatsHandler{
// 			To:  "classifier",
// 			Log: r.Log,
// 		}),
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	r.gRpcClientConn = classifierConn
//
// 	// Create client API for service
// 	classifierApi := proto.NewEventServiceClient(r.gRpcClientConn)
// 	return classifierApi, nil
// }
