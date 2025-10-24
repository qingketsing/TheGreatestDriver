async function fetchList(){
  const res = await fetch('/list');
  if(!res.ok) throw new Error('fetch list failed')
  const data = await res.json();
  const ul = document.getElementById('list');
  ul.innerHTML = '';
  data.forEach(item => {
    const li = document.createElement('li');
    const left = document.createElement('div');
    left.innerHTML = `<div><strong>${escapeHtml(item.things)}</strong></div><div class="meta">${escapeHtml(item.ddl)} • id:${item.id}</div>`;
    const del = document.createElement('button');
    del.textContent = 'Delete';
    del.addEventListener('click', async ()=>{
      if(!confirm('删除该条目?')) return;
      const resp = await fetch('/delete/'+item.id, {method:'DELETE'});
      if(resp.ok) fetchList(); else alert('删除失败');
    });
    li.appendChild(left);
    li.appendChild(del);
    ul.appendChild(li);
  });
}

function escapeHtml(s){
  if(!s) return '';
  return s.replace(/[&<>"']/g, c=>({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#39;"}[c]));
}

document.getElementById('addForm').addEventListener('submit', async (e)=>{
  e.preventDefault();
  const ddl = document.getElementById('ddl').value;
  const things = document.getElementById('things').value.trim();
  if(!ddl || !things) return alert('请填写 ddl 和 things');
  const body = new URLSearchParams();
  body.append('ddl', ddl);
  body.append('things', things);
  const res = await fetch('/append', {method:'POST', body});
  if(res.ok){
    document.getElementById('things').value = '';
    fetchList();
  } else {
    const text = await res.text();
    alert('Add failed: '+text);
  }
});

fetchList().catch(err=>{
  console.error(err);
  document.getElementById('list').innerHTML = '<li>加载失败: '+err.message+'</li>';
});