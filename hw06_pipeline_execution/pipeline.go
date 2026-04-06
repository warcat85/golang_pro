package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	numStages := len(stages)
	if numStages == 0 {
		return in
	}

	current := in
	for _, stage := range stages {
		current = stage(canceler(current, done))
	}
	// to make sure we don't return result if we were terminated
	current = canceler(current, done)
	return current
}

func canceler(in In, done In) Out {
	if done == nil {
		return in
	}
	out := make(Bi)
	go func() {
		defer close(out)

		for {
			select {
			case <-done:
				drain(in)
				return
			case v, ok := <-in:
				if !ok {
					return
				}

				select {
				case <-done:
					drain(in)
					return
				case out <- v:
				}
			}
		}
	}()
	return out
}

// drain all the values from the input channel before terminating.
func drain(in In) {
	for v := range in {
		_ = v
	}
}
