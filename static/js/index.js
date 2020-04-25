$(document).ready(function () {
    $("#file-upload").on('change', function () {
        url = "http://127.0.0.1:3001/file";
        file = $("#file-upload")[0].files[0];
        let data = new FormData();
        data.append("file", file);
        successFx = function (ans) {
            console.log("File successfully");
        };
        errorFx = function(ans){
            console.log("File failed");
        };
        $.ajax({
            contentType: false,
            processData: false,
            type: "POST",
            url: url,
            success: successFx,
            data : data
        })
    });
});