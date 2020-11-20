package tunnel

import "testing"

func TestServerRequest(t *testing.T) {
	s1, s2 := makePipe()

	s2.Listen("echo", func(ctx Context) {
		text, err := ctx.GetText()
		if err != nil {
			t.Error(err)
			return
		}
		ctx.Text(text)
	})

	t.Run("test request", func(t *testing.T) {
		text := "hello,world,there"
		resp, err := s1.SendText("echo", text)
		if err != nil {
			t.Error(err)
			return
		}
		if resp != text {
			t.Error("resp not equal text")
			return
		}
	})
}
