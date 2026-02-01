export interface FileDiff {
  oldPath: string;
  newPath: string;
  originalContent: string;
  modifiedContent: string;
  language: string;
}

interface HunkLine {
  type: 'context' | 'add' | 'delete';
  content: string;
}

interface Hunk {
  oldStart: number;
  oldCount: number;
  newStart: number;
  newCount: number;
  lines: HunkLine[];
}

const getLanguageFromPath = (filePath: string): string => {
  const ext = filePath.split('.').pop()?.toLowerCase() || '';
  const languageMap: Record<string, string> = {
    ts: 'typescript',
    tsx: 'typescript',
    js: 'javascript',
    jsx: 'javascript',
    py: 'python',
    rb: 'ruby',
    go: 'go',
    rs: 'rust',
    java: 'java',
    kt: 'kotlin',
    swift: 'swift',
    c: 'c',
    cpp: 'cpp',
    h: 'c',
    hpp: 'cpp',
    cs: 'csharp',
    php: 'php',
    html: 'html',
    css: 'css',
    scss: 'scss',
    less: 'less',
    json: 'json',
    yaml: 'yaml',
    yml: 'yaml',
    xml: 'xml',
    md: 'markdown',
    sql: 'sql',
    sh: 'shell',
    bash: 'shell',
    zsh: 'shell',
  };
  return languageMap[ext] || 'plaintext';
};

const parseHunkHeader = (line: string): Hunk | null => {
  const match = line.match(/^@@\s+-(\d+)(?:,(\d+))?\s+\+(\d+)(?:,(\d+))?\s+@@/);
  if (!match) return null;
  return {
    oldStart: parseInt(match[1], 10),
    oldCount: match[2] !== undefined ? parseInt(match[2], 10) : 1,
    newStart: parseInt(match[3], 10),
    newCount: match[4] !== undefined ? parseInt(match[4], 10) : 1,
    lines: [],
  };
};

const reconstructContent = (hunks: Hunk[], side: 'old' | 'new'): string => {
  if (hunks.length === 0) return '';

  const lines: string[] = [];
  let currentLine = 1;

  for (const hunk of hunks) {
    const targetStart = side === 'old' ? hunk.oldStart : hunk.newStart;

    while (currentLine < targetStart) {
      lines.push('');
      currentLine++;
    }

    for (const line of hunk.lines) {
      if (side === 'old') {
        if (line.type === 'context' || line.type === 'delete') {
          lines.push(line.content);
          currentLine++;
        }
      } else {
        if (line.type === 'context' || line.type === 'add') {
          lines.push(line.content);
          currentLine++;
        }
      }
    }
  }

  return lines.join('\n');
};

export const parseDiff = (diffText: string): FileDiff[] => {
  const files: FileDiff[] = [];
  const lines = diffText.split('\n');
  
  let currentFile: { oldPath: string; newPath: string; hunks: Hunk[] } | null = null;
  let currentHunk: Hunk | null = null;
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    if (line.startsWith('diff --git')) {
      if (currentFile && currentHunk) {
        currentFile.hunks.push(currentHunk);
      }
      if (currentFile) {
        files.push({
          oldPath: currentFile.oldPath,
          newPath: currentFile.newPath,
          originalContent: reconstructContent(currentFile.hunks, 'old'),
          modifiedContent: reconstructContent(currentFile.hunks, 'new'),
          language: getLanguageFromPath(currentFile.newPath || currentFile.oldPath),
        });
      }

      const pathMatch = line.match(/diff --git a\/(.+) b\/(.+)/);
      currentFile = {
        oldPath: pathMatch?.[1] || '',
        newPath: pathMatch?.[2] || '',
        hunks: [],
      };
      currentHunk = null;
      i++;
      continue;
    }

    if (line.startsWith('--- ')) {
      if (currentFile) {
        const path = line.slice(4).replace(/^a\//, '');
        if (path !== '/dev/null') currentFile.oldPath = path;
      }
      i++;
      continue;
    }

    if (line.startsWith('+++ ')) {
      if (currentFile) {
        const path = line.slice(4).replace(/^b\//, '');
        if (path !== '/dev/null') currentFile.newPath = path;
      }
      i++;
      continue;
    }

    if (line.startsWith('@@')) {
      if (currentHunk && currentFile) {
        currentFile.hunks.push(currentHunk);
      }
      currentHunk = parseHunkHeader(line);
      i++;
      continue;
    }

    if (currentHunk) {
      if (line.startsWith('+')) {
        currentHunk.lines.push({ type: 'add', content: line.slice(1) });
      } else if (line.startsWith('-')) {
        currentHunk.lines.push({ type: 'delete', content: line.slice(1) });
      } else if (line.startsWith(' ') || line === '') {
        currentHunk.lines.push({ type: 'context', content: line.slice(1) });
      }
    }

    i++;
  }

  if (currentFile && currentHunk) {
    currentFile.hunks.push(currentHunk);
  }
  if (currentFile) {
    files.push({
      oldPath: currentFile.oldPath,
      newPath: currentFile.newPath,
      originalContent: reconstructContent(currentFile.hunks, 'old'),
      modifiedContent: reconstructContent(currentFile.hunks, 'new'),
      language: getLanguageFromPath(currentFile.newPath || currentFile.oldPath),
    });
  }

  return files;
};
