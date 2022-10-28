import axios from "axios";

async function upload(file: File) {
    let form = new FormData()
    form.append("file", file)
    await axios.post("/objects/upload", form)
}

async function download(name: string) {
    let response = await axios.get(`/objects/download/${name}`, {
        responseType: 'blob', // important
    })
    // create file link in browser's memory
    const href = URL.createObjectURL(response.data);

    // create "a" HTML element with href to file & click
    const link = document.createElement('a');
    link.href = href;
    link.setAttribute('download', name);
    document.body.appendChild(link);
    link.click();

    // clean up "a" element & remove ObjectURL
    document.body.removeChild(link);
    URL.revokeObjectURL(href);
}

export {
    upload, download
}