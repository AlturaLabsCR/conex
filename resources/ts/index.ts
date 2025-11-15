export function toggleModal(event: Event) {
  event.preventDefault();

  const target = event.currentTarget as HTMLElement | null;
  if (!target) return;

  const modalId = target.dataset.target;
  if (!modalId) return;

  const modal = document.getElementById(modalId);
  if (modal) modal.classList.toggle('open');
}
(window as any).toggleModal = toggleModal;
