import { defineConfig } from 'vitest/config';

export default defineConfig({
	test: {
		include: ['src/**/*.{test,spec}.{js,ts}']
		// M-v0.2.1+ scope: 단위 테스트 추가 시 작성 (PR #663 별도)
	}
});
