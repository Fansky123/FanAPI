<template>
  <div class="docs-page">
    <div ref="scalarRoot" class="scalar-root" />
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'

const scalarRoot = ref(null)
let scriptEl = null

onMounted(() => {
  // 挂载 Scalar 容器元素
  const el = document.createElement('div')
  el.id = 'api-reference'
  el.setAttribute('data-url', '/api/swagger/doc.json')
  el.setAttribute(
    'data-configuration',
    JSON.stringify({ theme: 'default', darkMode: false, layout: 'sidebar', hideDarkModeToggle: true })
  )
  scalarRoot.value.appendChild(el)

  // 动态加载 Scalar CDN 脚本
  scriptEl = document.createElement('script')
  scriptEl.src = 'https://cdn.jsdelivr.net/npm/@scalar/api-reference'
  scriptEl.async = true
  document.head.appendChild(scriptEl)
})

onUnmounted(() => {
  if (scriptEl) scriptEl.remove()
})
</script>

<style scoped>
.docs-page {
  width: 100%;
  height: calc(100vh - 48px);
  overflow: hidden;
}
.scalar-root {
  width: 100%;
  height: 100%;
}
</style>
