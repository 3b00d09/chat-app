document.querySelector("body").addEventListener("htmx:beforeSwap", (e)=>{
  if(e.detail.xhr.status === 422){
    e.detail.shouldSwap = true;
    e.detail.isError = false;
  }
})

document.querySelector("#searchBtn").addEventListener("click",()=>{
    document.querySelector("#search-modal").showModal()
})

document.querySelector("#search-modal").addEventListener("click", (e) => {
  const modalContent = e.currentTarget.querySelector("div");
  if (!modalContent.contains(e.target)) {
    e.currentTarget.close();
  }
}); 

