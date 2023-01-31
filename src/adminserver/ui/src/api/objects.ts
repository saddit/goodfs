import axios from "axios";

async function upload(file: File, bucket: string) {
    let form = new FormData()
    form.append("file", file, file.name)
    form.append("bucket", bucket)
    await axios.put("/objects/upload", form, {
        timeout: 0
    })
}

async function download(name: string, bucket: string, version: number) {
    let response = await axios.get(`/objects/download/${name}?version=${version}&bucket=${bucket}`, {
        responseType: 'blob', // important
        timeout: 0
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

async function join(serverId: string) {
    await axios.post(`/objects/join/${serverId}`, {})
}

async function leave(serverId: string) {
    await axios.post(`/objects/leave/${serverId}`, {})
}

export {
    upload, download, join, leave
}