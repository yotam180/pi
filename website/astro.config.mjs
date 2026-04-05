// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightPageActions from 'starlight-page-actions';

export default defineConfig({
	integrations: [
		starlight({
			title: 'PI',
			description: 'Structured, polyglot, shareable developer automations',
			plugins: [
				starlightPageActions({
					actions: {
						chatgpt: true,
						claude: true,
						markdown: true,
					},
				}),
			],
			social: [
				{
					icon: 'github',
					label: 'GitHub',
					href: 'https://github.com/yotam180/pi',
				},
			],
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'getting-started/introduction' },
						{ label: 'Installation', slug: 'getting-started/installation' },
						{ label: 'Quick Start', slug: 'getting-started/quick-start' },
					],
				},
				{
					label: 'Concepts',
					items: [
						{ label: 'Automations', slug: 'concepts/automations' },
						{ label: 'pi.yaml', slug: 'concepts/pi-yaml' },
						{ label: 'Step Types', slug: 'concepts/step-types' },
						{ label: 'Shell Shortcuts', slug: 'concepts/shell-shortcuts' },
						{ label: 'Packages', slug: 'concepts/packages' },
					],
				},
				{
					label: 'Guides',
					items: [
						{ label: 'Setup Automations', slug: 'guides/setup-automations' },
						{ label: 'Cross-Platform Scripts', slug: 'guides/cross-platform-scripts' },
						{ label: 'Publishing to GitHub', slug: 'guides/publishing-to-github' },
						{ label: 'Using Packages', slug: 'guides/using-packages' },
						{ label: 'Private Repositories', slug: 'guides/private-repos' },
						{ label: 'Parent Shell Steps', slug: 'guides/parent-shell-steps' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ label: 'CLI Commands', slug: 'reference/cli' },
						{ label: 'Automation YAML', slug: 'reference/automation-yaml' },
						{ label: 'Conditions (if:)', slug: 'reference/conditions' },
						{ label: 'Built-in Automations', slug: 'reference/builtins' },
						{ label: 'pi-package.yaml', slug: 'reference/pi-package-yaml' },
					],
				},
			],
		}),
	],
});
