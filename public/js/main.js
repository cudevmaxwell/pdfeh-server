$(function() {
  $("select").change(function () {
    $(this).parentsUntil(".row-fluid").parent().find("select").val($(this).val());
  }).change();
});