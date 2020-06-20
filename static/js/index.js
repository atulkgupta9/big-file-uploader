$(document).ready(function () {
    $("#file-upload").on('change', function () {
        url = "http://127.0.0.1:3001/file";
        file = $("#file-upload")[0].files[0];
        let data = new FormData();
        data.append("file", file);
        successFx = function (ans) {
            let $errorA = $("#error-alert");
            $errorA.removeClass("alert-danger");
            $errorA.removeClass("alert-warning");
            $errorA.addClass("alert-success");
            let $ac = $("#ac");

            $ac.text("Successfully uploaded the file");
            $('#error-alert').show();

        };
        errorFx = function (ans) {
            let $ac = $("#ac");
            $("#addEmployeeModal").modal('toggle');
            $ac.text("Error while uploading the file");
            $('#error-alert').show();
            $errorA.removeClass("alert-warning");
            $errorA.removeClass("alert-success");
            $errorA.addClass("alert-danger");
            console.log("err", res);
        };
        $.ajax({
            contentType: false,
            processData: false,
            type: "POST",
            url: url,
            success: successFx,
            data: data,
            beforeSend: function(){
                let $errorA = $("#error-alert");
                $errorA.removeClass("alert-success");
                $errorA.removeClass("alert-danger");
                $errorA.addClass("alert-warning");

                let $ac = $("#ac");

                $ac.text("Uploading the file");
                $('#error-alert').show();

            }
        })
    });
});