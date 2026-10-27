import './style.css';
import { mount } from 'svelte';
import App from '$app/App.svelte';
import { registry } from '$app/modules';
import { setupI18n } from '$i18n/index';
import { applyTheme, getTheme } from '$shared/theme';

applyTheme(getTheme());
setupI18n(registry.modules);

const app = mount(App, { target: document.getElementById('app')! });

export default app;
