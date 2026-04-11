function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
function getAssistantNodes(){return Array.from(document.querySelectorAll('div.font-claude-response')).filter(el => !el.parentElement?.closest('div.font-claude-response'));}
function getLatestAssistantNode(){const nodes=getAssistantNodes(); return nodes.length ? nodes[nodes.length-1] : null;}
function hasCompletedAssistantActions(node){return !!node && !!node.querySelector('[data-testid="action-bar-copy"], [data-testid="action-bar-retry"], [aria-label="Copy"], [aria-label="Retry"], [aria-label="Edit"]');}
const node=getLatestAssistantNode();
return {
  href:location.href,
  title:document.title,
  hasNode: !!node,
  text: normalize(node?.innerText||node?.textContent),
  hasActions: hasCompletedAssistantActions(node),
  actionButtons: node ? Array.from(node.querySelectorAll('button')).map(el => ({text: normalize(el.innerText||el.textContent), aria: normalize(el.getAttribute('aria-label')), dataTestid: el.getAttribute('data-testid')})) : [],
};
