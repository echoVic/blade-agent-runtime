import { useState, useMemo } from 'react';
import { DiffEditor } from '@monaco-editor/react';
import { ChevronDown, ChevronRight, FileCode, FilePlus, FileMinus } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { parseDiff, type FileDiff } from '@/utils/parseDiff';

interface DiffViewerProps {
  diffContent: string;
}

interface FileDiffSectionProps {
  file: FileDiff;
  defaultExpanded?: boolean;
}

const getFileIcon = (oldPath: string, newPath: string) => {
  if (oldPath === '/dev/null' || !oldPath) {
    return <FilePlus className="w-4 h-4 text-emerald-400" />;
  }
  if (newPath === '/dev/null' || !newPath) {
    return <FileMinus className="w-4 h-4 text-rose-400" />;
  }
  return <FileCode className="w-4 h-4 text-blue-400" />;
};

const getFileName = (path: string): string => {
  return path.split('/').pop() || path;
};

const FileDiffSection = ({ file, defaultExpanded = true }: FileDiffSectionProps) => {
  const [expanded, setExpanded] = useState(defaultExpanded);
  const displayPath = file.newPath || file.oldPath;
  
  const lineCount = Math.max(
    file.originalContent.split('\n').length,
    file.modifiedContent.split('\n').length
  );
  const editorHeight = Math.min(Math.max(lineCount * 20 + 20, 150), 600);

  return (
    <div className="overflow-hidden rounded-lg border border-zinc-800 bg-zinc-950/50">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex gap-2 items-center px-4 py-3 w-full text-left transition-colors bg-zinc-900/80 hover:bg-zinc-800/80"
      >
        {expanded ? (
          <ChevronDown className="flex-shrink-0 w-4 h-4 text-zinc-400" />
        ) : (
          <ChevronRight className="flex-shrink-0 w-4 h-4 text-zinc-400" />
        )}
        {getFileIcon(file.oldPath, file.newPath)}
        <span className="text-sm font-medium truncate text-zinc-200" title={displayPath}>
          {getFileName(displayPath)}
        </span>
        <span className="ml-auto text-xs truncate text-zinc-500" title={displayPath}>
          {displayPath}
        </span>
      </button>

      <AnimatePresence initial={false}>
        {expanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="overflow-hidden"
          >
            <div style={{ height: editorHeight }} className="border-t border-zinc-800">
              <DiffEditor
                key={`${file.oldPath}-${file.newPath}-${file.originalContent.length}-${file.modifiedContent.length}`}
                original={file.originalContent}
                modified={file.modifiedContent}
                language={file.language}
                theme="vs-dark"
                options={{
                  readOnly: true,
                  renderSideBySide: true,
                  minimap: { enabled: false },
                  scrollBeyondLastLine: false,
                  fontSize: 13,
                  fontFamily: 'JetBrains Mono, Menlo, Monaco, monospace',
                  lineNumbers: 'on',
                  renderLineHighlight: 'none',
                  scrollbar: {
                    vertical: 'auto',
                    horizontal: 'auto',
                  },
                  diffWordWrap: 'off',
                  ignoreTrimWhitespace: false,
                  renderIndicators: true,
                  originalEditable: false,
                  hideUnchangedRegions: {
                    enabled: true,
                    revealLineCount: 3,
                    minimumLineCount: 5,
                    contextLineCount: 3,
                  },
                }}
              />
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

const DiffViewer = ({ diffContent }: DiffViewerProps) => {
  const fileDiffs = useMemo(() => parseDiff(diffContent), [diffContent]);

  if (fileDiffs.length === 0) {
    return (
      <div className="flex justify-center items-center h-full text-zinc-500">
        <p>No diff content to display</p>
      </div>
    );
  }

  return (
    <div className="flex overflow-y-auto flex-col gap-4 p-4 h-full">
      {fileDiffs.map((file, index) => (
        <FileDiffSection
          key={`${file.newPath || file.oldPath}-${index}`}
          file={file}
          defaultExpanded={index === 0}
        />
      ))}
    </div>
  );
};

export default DiffViewer;
