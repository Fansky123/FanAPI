<template>
  <div class="docs-page">
    <el-card class="hero-card">
      <div class="hero-row">
        <div>
          <div class="eyebrow">Developer Docs</div>
          <h3>接口文档与多语言示例</h3>
          <p>在控制台内直接查看 API 文档，包含 LLM、图像、视频、音频和任务轮询调用说明。</p>
        </div>
        <el-button type="primary" plain @click="openDocs">新窗口打开</el-button>
      </div>
    </el-card>
    <div class="docs-wrap">
      <iframe src="/api/docs" frameborder="0" class="docs-frame" />
    </div>
  </div>
</template>

<script setup>
function openDocs() {
  window.open('/api/docs', '_blank', 'noopener')
}
</script>

<style scoped>
.docs-page {
  max-width: 1320px;
}
.hero-card {
  margin-bottom: 16px;
}
.hero-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.eyebrow {
  color: #165dff;
  font-size: .82rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: .08em;
}
.hero-row h3 {
  margin: 8px 0 10px;
  font-size: 1.55rem;
}
.hero-row p {
  margin: 0;
  color: #617086;
}
.docs-wrap {
  height: calc(100vh - 210px);
  min-height: 720px;
  border-radius: 18px;
  overflow: hidden;
  border: 1px solid #dfe8f5;
  box-shadow: 0 14px 44px rgba(11, 33, 79, 0.08);
}
.docs-frame {
  width: 100%;
  height: 100%;
  border: none;
  display: block;
  background: #0f1117;
}

@media (max-width: 900px) {
  .hero-row {
    flex-direction: column;
    align-items: flex-start;
  }
  .docs-wrap {
    height: calc(100vh - 260px);
    min-height: 560px;
  }
}
</style>
