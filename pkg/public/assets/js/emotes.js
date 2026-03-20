// Emote shortcode to emoji map
// MOre entries to be added to expand support
const EMOTES = {
    ':smile:'         : '😊',
    ':grin:'          : '😁',
    ':laugh:'         : '😂',
    ':rofl:'          : '🤣',
    ':wink:'          : '😉',
    ':cool:'          : '😎',
    ':think:'         : '🤔',
    ':wow:'           : '😮',
    ':sad:'           : '😢',
    ':cry:'           : '😭',
    ':angry:'         : '😠',
    ':rage:'          : '😡',
    ':heart:'         : '❤️',
    ':broken_heart:'  : '💔',
    ':fire:'          : '🔥',
    ':100:'           : '💯',
    ':thumbsup:'      : '👍',
    ':thumbsdown:'    : '👎',
    ':wave:'          : '👋',
    ':clap:'          : '👏',
    ':eyes:'          : '👀',
    ':skull:'         : '💀',
    ':poop:'          : '💩',
    ':party:'         : '🎉',
    ':tada:'          : '🎊',
    ':star:'          : '⭐',
    ':sparkles:'      : '✨',
    ':check:'         : '✅',
    ':x:'             : '❌',
    ':warning:'       : '⚠️',
    ':bench:'         : '🪑',
    ':lock:'          : '🔒',
    ':motonai:'       : '🦓',
    ':izi:'           : '🐻',
};

//This function replaces all :shortcode: patterns in a string with their emoji.
//Called after escapeHtml() so shortcodes survive HTML escaping. :[\w]+: = word
function applyEmotes(text) {
    return text.replace(/:[\w]+:/g, (match) => EMOTES[match] || match);
}
