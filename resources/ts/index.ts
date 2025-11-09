export function toggleModal(event) {
  event.preventDefault();
  const modal = document.getElementById(event.currentTarget.dataset.target);
  if (modal) modal.classList.toggle('open');
}

window.toggleModal = toggleModal;
