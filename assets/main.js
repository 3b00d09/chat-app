document.querySelector("body").addEventListener("htmx:beforeSwap", (e)=>{
  if(e.detail.xhr.status === 422){
    e.detail.shouldSwap = true;
    e.detail.isError = false;
  }
})

// scroll to end after a swap in the messages container, cant make htmx scroll bottom work
document.body.addEventListener("htmx:oobAfterSwap", (e)=>{
  if(e.detail.target.id === "messages"){
    document.querySelector("#messages_container").scrollTo(0, document.querySelector("#messages_container").scrollHeight);
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

