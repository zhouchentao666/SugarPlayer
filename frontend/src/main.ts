import '@wailsio/runtime'
import { createApp } from 'vue'
import App from './App.vue'
import EditorApp from './EditorApp.vue'
import './style.css'

const params = new URLSearchParams(window.location.search)
const root = params.get('editor') === '1' ? EditorApp : App

createApp(root).mount('#app')
