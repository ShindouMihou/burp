package docker

import "github.com/docker/docker/errdefs"

func nonReturningTask(task func() error) error {
	_, err := returningTask(func() (*any, error) { return nil, task() })
	return err
}

func returningTask[T any](task func() (T, error)) (*T, error) {
	t, err := task()
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func has(task func() (any, error)) (bool, error) {
	res, err := returningTask(task)
	if err != nil {
		return false, err
	}
	return res != nil, nil
}
