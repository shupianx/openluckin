<script setup>
import { computed, ref } from 'vue'
import CoffeeCup from './components/CoffeeCup.vue'

const origin = window.location.origin.includes('localhost')
  ? 'https://openluckin.com' // 本地预览时回退到正式域名
  : window.location.origin

const zipUrl = `${origin}/openluckin-order.zip`
const copyText = `请下载安装 OpenLuckin Skill：\n${zipUrl}`
// GitHub 即注册表：npx skills 会扫描仓库内的 SKILL.md 并装到本机各 agent
const quickCmd = 'npx skills add shupianx/openluckin'

// 手动安装 CLI：按平台切换命令
const cliCmds = {
  mac: `curl -fsSL ${origin}/install.sh | bash`,
  win: `irm ${origin}/install.ps1 | iex`,
}
const platform = ref('mac')
const ddOpen = ref(false)
function pickPlatform(p) {
  platform.value = p
  ddOpen.value = false
}
// 命令区按最长命令定宽，切换平台时盒子宽度不跳变（等宽字体下 1 字符 = 1ch）
const cliMinWidth = computed(
  () => Math.max(...Object.values(cliCmds).map((c) => c.length)) + 'ch'
)

const copied = ref('')
async function copy(text, key) {
  await navigator.clipboard.writeText(text)
  copied.value = key
  setTimeout(() => (copied.value = ''), 1500)
}
</script>

<template>
  <header class="hero">
    <div class="site-title">
      <img class="site-logo" src="/logo-light.svg" alt="OpenLuckin logo" />
      <span>OpenLuckin</span>
    </div>
    <a class="github-link" href="https://github.com/shupianx/openluckin" target="_blank" rel="noopener"
      aria-label="GitHub" title="GitHub">
      <svg width="26" height="26" viewBox="0 0 16 16" fill="currentColor">
        <path
          d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8z" />
      </svg>
    </a>
    <p class="slogan">一句话，幸运到手</p>
    <div class="hero-main">
      <div class="hero-left">
        <CoffeeCup />
      </div>
      <div class="get-skill">
      <p class="intro">
        OpenLuckin 是非官方瑞幸点单工具（CLI + Agent
        Skill），基于瑞幸官方接口封装。安装后对你的智能体说一句「帮我点杯生椰拿铁」，
        它就能完成找店、点单、付款、报取餐码的全流程。
      </p>
      <p class="cmd-label">快速安装：</p>
      <div class="cmd quick-cmd">
        <code>{{ quickCmd }}</code>
        <button class="copy-btn" :class="{ done: copied === 'quick' }" aria-label="复制" title="复制"
          @click="copy(quickCmd, 'quick')">
          <svg v-if="copied !== 'quick'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
            stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
            <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
          </svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"
            stroke-linecap="round" stroke-linejoin="round">
            <path d="M20 6 9 17l-5-5" />
          </svg>
        </button>
      </div>

      <p class="cmd-label cli-label">复制下面内容发给你的智能体（OpenClaw、Hermes 等）：</p>
      <div class="cmd">
        <code>请下载安装 OpenLuckin Skill：<br />{{ zipUrl }}</code>
      <button class="copy-btn" :class="{ done: copied === 'skill' }" aria-label="复制" title="复制"
        @click="copy(copyText, 'skill')">
        <svg v-if="copied !== 'skill'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
          <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
        </svg>
        <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"
          stroke-linecap="round" stroke-linejoin="round">
          <path d="M20 6 9 17l-5-5" />
        </svg>
        </button>
      </div>

      <p class="cmd-label cli-label">手动安装CLI（可选）</p>
      <div class="cmd cli-cmd">
        <div class="dropdown">
          <button class="dd-trigger" @click="ddOpen = !ddOpen">
            <svg v-if="platform === 'mac'" width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
              <path
                d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47c-1.34.03-1.77-.79-3.29-.79c-1.53 0-2 .77-3.27.82c-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51c1.28-.02 2.5.87 3.29.87c.78 0 2.26-1.07 3.81-.91c.65.03 2.47.26 3.64 1.98c-.09.06-2.17 1.28-2.15 3.81c.03 3.02 2.65 4.03 2.68 4.04c-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5c.13 1.17-.34 2.35-1.04 3.19c-.69.85-1.83 1.51-2.95 1.42c-.15-1.15.41-2.35 1.05-3.11" />
            </svg>
            <svg v-else width="16" height="16" viewBox="0 0 256 256">
              <path fill="#F1511B" d="M121.666 121.666H0V0h121.666z" />
              <path fill="#80CC28" d="M256 121.666H134.335V0H256z" />
              <path fill="#00ADEF" d="M121.663 256.002H0V134.336h121.663z" />
              <path fill="#FBBC09" d="M256 256.002H134.335V134.336H256z" />
            </svg>
            <span>{{ platform === 'mac' ? 'macOS' : 'Windows' }}</span>
            <svg class="dd-caret" :class="{ open: ddOpen }" width="12" height="12" viewBox="0 0 24 24" fill="none"
              stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="m6 9 6 6 6-6" />
            </svg>
          </button>
          <ul v-if="ddOpen" class="dd-menu">
            <li @click="pickPlatform('mac')">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                <path
                  d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47c-1.34.03-1.77-.79-3.29-.79c-1.53 0-2 .77-3.27.82c-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51c1.28-.02 2.5.87 3.29.87c.78 0 2.26-1.07 3.81-.91c.65.03 2.47.26 3.64 1.98c-.09.06-2.17 1.28-2.15 3.81c.03 3.02 2.65 4.03 2.68 4.04c-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5c.13 1.17-.34 2.35-1.04 3.19c-.69.85-1.83 1.51-2.95 1.42c-.15-1.15.41-2.35 1.05-3.11" />
              </svg>
              <span>macOS</span>
            </li>
            <li @click="pickPlatform('win')">
              <svg width="16" height="16" viewBox="0 0 256 256">
                <path fill="#F1511B" d="M121.666 121.666H0V0h121.666z" />
                <path fill="#80CC28" d="M256 121.666H134.335V0H256z" />
                <path fill="#00ADEF" d="M121.663 256.002H0V134.336h121.663z" />
                <path fill="#FBBC09" d="M256 256.002H134.335V134.336H256z" />
              </svg>
              <span>Windows</span>
            </li>
          </ul>
        </div>
        <code :style="{ minWidth: cliMinWidth }">{{ cliCmds[platform] }}</code>
        <button class="copy-btn" :class="{ done: copied === 'cli' }" aria-label="复制" title="复制"
          @click="copy(cliCmds[platform], 'cli')">
          <svg v-if="copied !== 'cli'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
            stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
            <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
          </svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"
            stroke-linecap="round" stroke-linejoin="round">
            <path d="M20 6 9 17l-5-5" />
          </svg>
        </button>
        </div>
      </div>
    </div>
  </header>

  <footer>
    <p>openluckin 是个人开源项目，与瑞幸咖啡官方无关。下单产生真实消费，请谨慎操作。</p>
  </footer>
</template>

<style scoped>
.hero {
  flex: 1; /* 占满页脚以上的全部视口高度 */
  position: relative; /* 站名绝对定位的锚点 */
  text-align: center;
  padding: 56px 24px 64px;
  /* 渐变须在 3D 画布上沿（约 150px）之前完成过渡，其下为纯品牌蓝，
     与画布内的纯色场景背景天然无缝（画布背景同为 #212370） */
  background: linear-gradient(180deg, #131542 0%, #212370 140px);
  color: #fff;
}
.site-title {
  position: absolute;
  top: 20px;
  left: 28px;
  display: flex;
  align-items: center;
  gap: 10px;
  font-family: 'Bitcount Single', ui-monospace, monospace;
  font-size: 26px;
  font-weight: 400;
  color: #fff;
  letter-spacing: 1px;
}
.site-logo {
  height: 26px; /* logo 宽高比约 2:1，随高度自适应 */
  width: auto;
}
.github-link {
  position: absolute;
  top: 22px;
  right: 28px;
  color: #fff;
  display: inline-flex;
}
.github-link:hover {
  opacity: 0.75;
}
/* 左右双栏：宽度足够时 slogan+3D 在左、复制框组在右；
   容器装不下两栏时 flex-wrap 自动折回上下结构 */
.hero-main {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 16px 56px;
  max-width: 1280px;
  margin: 0 auto;
}
.hero-left {
  flex: 1 1 520px;
  max-width: 760px;
  min-width: 0; /* 允许画布随栏宽收缩 */
}
.slogan {
  font-size: 32px;
  font-weight: 600;
  color: #fff;
  letter-spacing: 6px;
  margin: 36px auto 12px; /* 上边距避开绝对定位的顶栏 */
}
.get-skill {
  /* 收缩到内容宽度，内部左对齐——说明文字与框左缘对齐 */
  flex: 0 0 auto;
  text-align: left;
}
.intro {
  font-size: 15px;
  line-height: 1.9;
  color: #c7cbf0;
  margin: 0 0 22px;
  /* 不参与列宽计算（否则长句会把右栏撑爆、挤垮左右布局），
     宽度跟随由复制框决定的列宽自动折行 */
  width: 0;
  min-width: 100%;
}
.cmd-label {
  font-size: 14px;
  color: #93a7e0;
  margin: 0 0 10px;
}
.cli-label {
  margin-top: 22px;
}
/* 双类提高优先级：覆盖 .cmd 的 flex-start（那是给 skill 框按钮贴右上角用的），
   单行内容的框全部垂直居中 */
.cmd.cli-cmd,
.cmd.quick-cmd {
  position: relative; /* 下拉菜单的定位锚点 */
  display: flex;
  align-items: center;
}
.cmd.cli-cmd .copy-btn,
.cmd.quick-cmd .copy-btn {
  margin: 0 -6px 0 0; /* 取消贴右上角的负边距，保持垂直居中 */
}
.cmd.quick-cmd {
  display: inline-flex; /* 自然宽度：框随内容收缩，不撑满容器 */
}
.cli-cmd code {
  display: inline-block; /* 让 min-width 生效，平台切换不引起宽度跳变 */
}
.dd-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between; /* 图标靠左、箭头钉右，文字差量被内部吸收 */
  gap: 7px;
  min-width: 124px; /* 锁定宽度 ≥ 最宽选项（Windows），切换平台不再引起框宽突变 */
  border: none;
  border-right: 1px solid rgba(255, 255, 255, 0.2); /* 与命令区的分隔线 */
  border-radius: 0;
  background: transparent;
  color: #fff;
  font-size: 14px;
  padding: 2px 12px 2px 0;
  margin: -2px 0; /* 让分隔线略高于文字 */
  cursor: pointer;
}
.dd-trigger:hover {
  color: #c9d6ff;
}
.dd-caret {
  transition: transform 0.15s;
  color: #93a7e0;
}
.dd-caret.open {
  transform: rotate(180deg);
}
.dd-menu {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  margin: 0;
  padding: 6px;
  list-style: none;
  background: #1a1c5a;
  border: 1px solid rgba(255, 255, 255, 0.22);
  border-radius: 10px;
  z-index: 10;
}
.dd-menu li {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 14px;
  white-space: nowrap;
  cursor: pointer;
}
.dd-menu li:hover {
  background: rgba(255, 255, 255, 0.12);
}
.cmd {
  display: inline-flex;
  align-items: flex-start; /* 复制按钮贴右上角 */
  gap: 14px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.22);
  color: #fff;
  border-radius: 10px;
  padding: 14px 18px;
  backdrop-filter: blur(4px);
  text-align: left;
}
.cmd code {
  font-size: 14px;
  line-height: 1.7;
}
.copy-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: #bcd0ff;
  padding: 6px;
  margin: -7px -10px 0 0; /* 抵消容器内边距，贴近右上角 */
  cursor: pointer;
}
.copy-btn:hover {
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
}
.copy-btn.done {
  color: #7ee2a8;
}

footer {
  text-align: center;
  color: rgba(255, 255, 255, 0.45);
  font-size: 13px;
  padding: 20px 24px;
}
</style>
