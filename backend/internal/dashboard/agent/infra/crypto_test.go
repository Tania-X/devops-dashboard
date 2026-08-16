package infra

import "testing"

func TestCrypto_RoundTrip(t *testing.T) {
	c := NewCrypto("test-secret-key")

	t.Run("加密解密往返", func(t *testing.T) {
		plain := "P@ssw0rd-2026"
		enc, err := c.Encrypt(plain)
		if err != nil {
			t.Fatalf("Encrypt 失败: %v", err)
		}
		if enc == plain {
			t.Error("密文不应等于明文")
		}
		dec, err := c.Decrypt(enc)
		if err != nil {
			t.Fatalf("Decrypt 失败: %v", err)
		}
		if dec != plain {
			t.Errorf("解密结果 = %q, want %q", dec, plain)
		}
	})

	t.Run("每次加密随机 nonce(密文不同)", func(t *testing.T) {
		e1, _ := c.Encrypt("same")
		e2, _ := c.Encrypt("same")
		if e1 == e2 {
			t.Error("相同明文两次加密应产生不同密文(随机 nonce)")
		}
	})

	t.Run("空密码解密报错", func(t *testing.T) {
		if _, err := c.Decrypt(""); err == nil {
			t.Error("空密文应报错")
		}
	})

	t.Run("短 key 自动补齐,长 key 截断", func(t *testing.T) {
		short := NewCrypto("x") // 1 字节 → pad 到 32
		enc, err := short.Encrypt("a")
		if err != nil {
			t.Fatalf("短 key 加密失败: %v", err)
		}
		if _, err := short.Decrypt(enc); err != nil {
			t.Errorf("短 key 解密失败: %v", err)
		}
		_ = NewCrypto("a-very-long-secret-key-that-exceeds-thirty-two-bytes-123456") // 截断不 panic
	})
}
