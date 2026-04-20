import os
import re
import textwrap

def fix_md001(lines):
    """Fix heading increments (MD001) while maintaining sibling consistency."""
    new_lines = []
    in_code_block = False
    
    # First pass: identify all heading levels and their intended hierarchy
    # We'll map each level to a new level that doesn't skip
    level_map = {0: 0}
    current_new_level = 0
    
    # We need to handle the headings dynamically
    # If we see a heading level L, and we know its parent level P (the last level < L),
    # then new_L = map[P] + 1.
    
    last_levels = {0: 0} # map of original_level -> last_new_level seen at that or higher level
    
    headings = []
    for line in lines:
        strip_line = line.strip()
        if strip_line.startswith('```'):
            in_code_block = not in_code_block
            continue
        if in_code_block:
            continue
        match = re.match(r'^(#+)\s+(.*)$', line)
        if match:
            headings.append(len(match.group(1)))
    
    # Create mapping
    mapping = {}
    current_hierarchy = [0] # Stack of (original, new)
    
    # This is complex. Let's simplify:
    # Most of these files are: # Title, then ### Tasks, then maybe ## Task.
    # They should all be ##.
    
    # Simpler approach: if level > prev_level + 1, then new_level = prev_level + 1.
    # To keep siblings consistent, if we have [1, 3, 3, 2], we want [1, 2, 2, 2].
    
    # Let's just track the last seen level and if we jump, we set a 'decrement' for that level.
    decrements = {}
    last_new_level = 0
    
    # Actually, let's just use the "no skip" rule and see if it's good enough.
    # To handle siblings, if we see a level L that we previously mapped to L', 
    # and the current new level is compatible, we keep it.
    
    return None # Will implement below

def wrap_text(text, width=80, indent=''):
    """Wrap text while preserving URLs and list markers."""
    if not text.strip():
        return text
    
    # If it's a list item, extract the marker
    list_match = re.match(r'^(\s*[-*+]|\s*\d+\.)\s+(.*)$', text)
    if list_match:
        marker = list_match.group(1)
        content = list_match.group(2)
        prefix = marker + ' '
        # For subsequent lines, indent by the length of the marker + 1
        sub_indent = ' ' * len(prefix)
    else:
        prefix = ''
        sub_indent = ''
    
    # Use textwrap but avoid breaking URLs
    # A simple way: split by space, and build lines
    words = (prefix + (content if list_match else text)).split(' ')
    lines = []
    current_line = []
    current_len = 0
    
    for word in words:
        word_len = len(word)
        # If word is a URL, don't break it
        if 'http' in word:
            if current_line and current_len + 1 + word_len > width:
                lines.append(' '.join(current_line))
                current_line = [sub_indent + word if lines else word]
                current_len = len(current_line[0])
            else:
                current_line.append(word)
                current_len += (1 if current_len > 0 else 0) + word_len
        else:
            if current_line and current_len + 1 + word_len > width:
                lines.append(' '.join(current_line))
                current_line = [sub_indent + word]
                current_len = len(current_line[0])
            else:
                current_line.append(word)
                current_len += (1 if current_len > 0 else 0) + word_len
                
    if current_line:
        lines.append(' '.join(current_line))
        
    return '\n'.join(lines) + '\n'

def process_file(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    # Pass 1: Fix MD001 (Heading levels)
    # Strategy: Just ensure no jumps. If jump, cap it. 
    # To keep siblings consistent, if we are in a "jumped" state, keep it until we hit a higher level.
    
    md001_lines = []
    current_level = 0
    in_code_block = False
    cap = 0
    
    for line in lines:
        strip_line = line.strip()
        if strip_line.startswith('```'):
            in_code_block = not in_code_block
            md001_lines.append(line)
            continue
        if in_code_block:
            md001_lines.append(line)
            continue
            
        match = re.match(r'^(#+)\s+(.*)$', line)
        if match:
            hashes = match.group(1)
            title = match.group(2)
            original_level = len(hashes)
            
            if current_level == 0:
                current_level = original_level
                md001_lines.append(line)
            else:
                if original_level > current_level + 1:
                    # Jump detected
                    new_level = current_level + 1
                    md001_lines.append(f"{'#' * new_level} {title}\n")
                    # We don't update current_level to original_level, 
                    # we update it to new_level to catch subsequent jumps of same original level
                    current_level = new_level
                else:
                    md001_lines.append(line)
                    current_level = original_level
        else:
            md001_lines.append(line)

    # Pass 2: Fix MD013 (Line length)
    final_lines = []
    in_code_block = False
    for line in md001_lines:
        strip_line = line.strip()
        if strip_line.startswith('```'):
            in_code_block = not in_code_block
            final_lines.append(line)
            continue
        if in_code_block or strip_line.startswith('#') or not strip_line:
            final_lines.append(line)
            continue
        
        # Only wrap if longer than 80
        if len(line) > 81:
            final_lines.append(wrap_text(line.rstrip('\n')))
        else:
            final_lines.append(line)

    with open(file_path, 'w', encoding='utf-8') as f:
        f.writelines(final_lines)

if __name__ == "__main__":
    conductor_dir = 'conductor'
    for filename in os.listdir(conductor_dir):
        if filename.endswith('.md'):
            process_file(os.path.join(conductor_dir, filename))
