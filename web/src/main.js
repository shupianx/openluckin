import { ViteSSG } from 'vite-ssg/single-page'
import './style.css'
import App from './App.vue'

// 单页模式：构建时用 Vue SSR 预渲染出真实 HTML，客户端再 hydrate。
// 导出名必须是 createApp，vite-ssg 的构建端据此找到应用工厂。
export const createApp = ViteSSG(App)
