export interface DiffLine {
  type: 'add' | 'remove' | 'context';
  line_num: number;
  content: string;
}

export interface DiffHunk {
  old_start: number;
  old_count: number;
  new_start: number;
  new_count: number;
  lines: DiffLine[];
}

export interface DiffResponse {
  format: string;
  hunks: DiffHunk[];
  added_lines: number;
  removed_lines: number;
  text: string;
}
