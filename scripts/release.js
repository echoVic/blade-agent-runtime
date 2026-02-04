#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const CHANGELOG_FILE = 'CHANGELOG.md';
const VERSION_FILE = 'cmd/bar/update.go';

function exec(cmd, options = {}) {
    try {
        return execSync(cmd, { encoding: 'utf8', ...options }).trim();
    } catch (e) {
        if (options.allowFailure) return null;
        throw e;
    }
}

function getCurrentVersion() {
    const content = fs.readFileSync(VERSION_FILE, 'utf8');
    const match = content.match(/currentVersion\s*=\s*"([^"]+)"/);
    return match ? match[1] : '0.0.0';
}

function bumpVersion(version, type) {
    const [major, minor, patch] = version.split('.').map(Number);
    switch (type) {
        case 'major': return `${major + 1}.0.0`;
        case 'minor': return `${major}.${minor + 1}.0`;
        case 'patch': return `${major}.${minor}.${patch + 1}`;
        default: return version;
    }
}

function updateVersion(version) {
    console.log(`==> Updating version to ${version}`);
    let content = fs.readFileSync(VERSION_FILE, 'utf8');
    content = content.replace(/currentVersion\s*=\s*"[^"]+"/, `currentVersion = "${version}"`);
    fs.writeFileSync(VERSION_FILE, content);
}

function generateChangelog(version) {
    const lastTag = exec('git describe --tags --abbrev=0', { allowFailure: true }) || '';
    const range = lastTag ? `${lastTag}..HEAD` : 'HEAD~20..HEAD';
    const commits = exec(`git log ${range} --pretty=format:"%s"`, { allowFailure: true }) || '';
    
    if (!commits) {
        console.log('Warning: No commits found');
        return '\n### Changed\n- Version bump\n';
    }

    const categories = {
        feat: { label: 'Added', items: [] },
        fix: { label: 'Fixed', items: [] },
        docs: { label: 'Documentation', items: [] },
        chore: { label: 'Chore', items: [] },
        refactor: { label: 'Changed', items: [] },
        perf: { label: 'Changed', items: [] },
        style: { label: 'Changed', items: [] },
    };

    commits.split('\n').forEach(line => {
        const match = line.match(/^(\w+)(?:\([^)]+\))?:\s*(.+)$/);
        if (match) {
            const [, type, msg] = match;
            if (categories[type]) {
                categories[type].items.push(msg);
            }
        }
    });

    const sections = {};
    Object.entries(categories).forEach(([, { label, items }]) => {
        if (items.length > 0) {
            if (!sections[label]) sections[label] = [];
            sections[label].push(...items);
        }
    });

    let content = '';
    Object.entries(sections).forEach(([label, items]) => {
        content += `\n### ${label}\n`;
        items.forEach(item => {
            content += `- ${item}\n`;
        });
    });

    return content || '\n### Changed\n- Version bump\n';
}

function updateChangelog(version) {
    console.log(`==> Generating CHANGELOG for ${version}`);
    
    const date = new Date().toISOString().split('T')[0];
    const changelogContent = generateChangelog(version);
    
    let changelog = fs.readFileSync(CHANGELOG_FILE, 'utf8');
    const lines = changelog.split('\n');
    
    const newEntry = `## [${version}] - ${date}\n${changelogContent}\n`;
    lines.splice(6, 0, '', newEntry);
    
    fs.writeFileSync(CHANGELOG_FILE, lines.join('\n'));
}

function checkCleanWorkingTree() {
    const status = exec('git status --porcelain');
    if (status) {
        console.error('Error: Working tree has uncommitted changes');
        console.error(status);
        process.exit(1);
    }
}

function prepareRelease(version) {
    checkCleanWorkingTree();
    
    const current = getCurrentVersion();
    console.log(`Current version: ${current}`);
    console.log(`New version: ${version}`);
    console.log('');
    
    updateVersion(version);
    updateChangelog(version);
    
    console.log('');
    console.log('==> Committing changes...');
    exec(`git add ${VERSION_FILE} ${CHANGELOG_FILE}`);
    exec(`git commit -m "chore: release v${version}"`);
    
    console.log('');
    console.log(`✓ Release ${version} prepared`);
}

function publishRelease() {
    checkCleanWorkingTree();
    
    const version = getCurrentVersion();
    const tag = `v${version}`;
    
    console.log(`==> Publishing version ${version}`);
    
    console.log('==> Building binaries...');
    exec('mkdir -p dist');
    
    const platforms = [
        { os: 'darwin', arch: 'amd64' },
        { os: 'darwin', arch: 'arm64' },
        { os: 'linux', arch: 'amd64' },
        { os: 'linux', arch: 'arm64' },
    ];
    
    platforms.forEach(({ os, arch }) => {
        console.log(`    Building ${os}/${arch}...`);
        exec(`GOOS=${os} GOARCH=${arch} go build -o dist/bar ./cmd/bar`);
        exec(`tar -czf dist/bar_${os}_${arch}.tar.gz -C dist bar`);
        exec('rm dist/bar');
    });
    
    console.log(`==> Creating tag ${tag}...`);
    exec(`git tag ${tag}`);
    
    console.log('==> Pushing to origin...');
    exec('git push');
    exec(`git push origin ${tag}`);
    
    console.log('==> Creating GitHub release...');
    exec(`gh release create ${tag} dist/bar_*.tar.gz --title "${tag}" --generate-notes`);
    
    console.log('');
    console.log(`✓ Released ${tag}`);
    console.log(`  https://github.com/echoVic/blade-agent-runtime/releases/tag/${tag}`);
}

function showHelp() {
    console.log(`Usage: node scripts/release.js <command>

Commands:
  patch              Bump patch version and release (0.0.1 -> 0.0.2)
  minor              Bump minor version and release (0.0.1 -> 0.1.0)
  major              Bump major version and release (0.0.1 -> 1.0.0)
  prepare <version>  Prepare release with specific version
  publish            Build and publish the release

Examples:
  node scripts/release.js patch
  node scripts/release.js minor
  node scripts/release.js prepare 0.0.15
  node scripts/release.js publish`);
}

const [, , command, arg] = process.argv;

switch (command) {
    case 'patch':
    case 'minor':
    case 'major':
        const current = getCurrentVersion();
        const newVersion = bumpVersion(current, command);
        prepareRelease(newVersion);
        publishRelease();
        break;
    case 'prepare':
        if (!arg) {
            console.error('Error: Version required');
            process.exit(1);
        }
        prepareRelease(arg);
        break;
    case 'publish':
        publishRelease();
        break;
    case '-h':
    case '--help':
    case 'help':
        showHelp();
        break;
    default:
        showHelp();
        process.exit(1);
}
