package service

type ollamaService struct {
	config map[string]string
}

func (d *ollamaService) Completion(prompt string) (*Result, error) {
	r := &Result{
		Completion: "hello world",
	}

	return r, nil
}
