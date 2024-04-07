document.querySelector("#searchBtn").addEventListener("click",()=>{
    document.querySelector("#search-modal").showModal()
})

document.querySelector("#search-modal").addEventListener("click", (e) => {
  const modalContent = e.currentTarget.querySelector("div");
  if (!modalContent.contains(e.target)) {
    e.currentTarget.close();
  }
}); 