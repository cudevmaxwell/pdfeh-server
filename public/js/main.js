$(function() {
  //Editor
  $("select").change(function () {
    $(this).parentsUntil(".row-fluid").parent().find("select").val($(this).val());
  }).change();
  
  //Inline Example  
  $('object[type="application/pdf"]').after('<span class="label label-info">Checking...</span>');  
  $('object[type="application/pdf"]').each(function() {
      var objChecked = $(this);
      var url = $(this).attr('data');
	  $.get("http://192.168.0.77/api", { "pdf": url, "validator": "http://192.168.0.77/public/examples/warn.json" })
	  .done(function(data) {
	    if (data.Level == "pass"){
		  objChecked.next().replaceWith('<span class="label label-success">PASSED!</span>'); 
		}
		else if (data.Level == "warn"){
		  objChecked.next().replaceWith('<span class="label label-warning">WARNING!</span>'); 
		}
		else if (data.Level == "fail"){
		  objChecked.next().replaceWith('<span class="label label-important">FAILED!</span>'); 
		}
        
      });
     
  });  
  
});