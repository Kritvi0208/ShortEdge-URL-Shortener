window.onload = async function () {
  const container = document.getElementById('allLinks');
  const origin = window.location.origin;
  try {
    const res = await fetch('/all');
    const links = await res.json();

    if (links.length === 0) {
      container.innerHTML = '<p>No links found.</p>';
      return;
    }

    container.innerHTML = ""; // Clear previous content

    links.forEach(link => {
      const div = document.createElement('div');
      div.className = 'link-card';

      div.innerHTML = `
        <p><strong>Original:</strong>
          <div class="scroll-url">
            <a href="${link.long_url}" target="_blank">${link.long_url}</a>
          </div>
        </p>
        <p><strong>Short:</strong>
          <div class="scroll-url">
            <a href="${origin}/r/${link.code}" target="_blank">${origin}/r/${link.code}</a>
          </div>
        </p>
        <p><strong>Visibility:</strong> ${link.visibility}</p>
      `;
      container.appendChild(div);
    });
  } catch (err) {
    container.innerHTML = '<p>Error loading links.</p>';
  }
};
