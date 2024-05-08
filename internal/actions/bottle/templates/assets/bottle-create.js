//global listeners
var isFormDirty = false;
window.addEventListener('beforeunload', (event) => {
  if (isFormDirty) {
    const confirmationMessage = "You have unsaved changes on your bottle that have not been saved.  Please click 'Save' at the bottom of the form to prevent losing this data.";
    event.returnValue = confirmationMessage;
    return confirmationMessage;
  }
});

document
.getElementById("bottle-form")
.addEventListener("submit", onSubmit, true);

document
.getElementById("discard-button")
.addEventListener("click", onDiscard, true);


/// define variables for inputs on the page to be referenced later
var descriptionText = document.getElementById("description");
var authorName = document.getElementById("author-name");
var authorEmail = document.getElementById("author-email");
var authorUrl = document.getElementById("author-url");
var authorTable = document
  .getElementById("author-table")
  .getElementsByTagName("tbody")[0];

var partsTable = document
  .getElementById("parts-table")
  .getElementsByTagName("tbody")[0];

var sourceTable = document
  .getElementById("source-table")
  .getElementsByTagName("tbody")[0];

var sourceName = document.getElementById("sourceName");
var sourceUrl = document.getElementById("sourceUri");
var labelList = document.getElementsByClassName("pills labels")[0];
var labelName = document.getElementById("label-name");
var labelValue = document.getElementById("label-value");

var annotationList = document.getElementsByClassName("pills annotations")[0];
var annotationName = document.getElementById("annotation-name");
var annotationValue = document.getElementById("annotation-value");

var metriclist = document.getElementsByClassName("pills metrics")[0];
var metricname = document.getElementById("metric-name");
var metricvalue = document.getElementById("metric-value");

var metricDesc = document.getElementById("metric-desc");

var partName = document.getElementById("partName");
var partValue = document.getElementById("partValue");

var formDiv = document.getElementById("bottle-form");
var submitErrorDiv = document.getElementById("submitError");

//form button elements by default "Submit" a postback, this is capturing that and 
//intercepting it to be handled client-side
function onSubmit(e) {
  e.preventDefault();
  isFormDirty = true;
  if (e.submitter.innerHTML.indexOf("ADD AUTHOR") > -1) {
    addAuthor(e);
  } else if (e.submitter.innerHTML.indexOf("SAVE AUTHOR") > -1) {
    saveAuthor(e);
  } else if (e.submitter.innerHTML.indexOf("ADD LABEL") > -1) {
    saveLabel(e);
  } else if (e.submitter.innerHTML.indexOf("ADD SOURCE") > -1) {
    addSource(e);
  } else if (e.submitter.innerHTML.indexOf("SAVE SOURCE") > -1) {
    saveSource(e);
  } else if (e.submitter.innerHTML.indexOf("ADD ANNOTATION") > -1) {
    saveAnnotation(e);
  } else if (e.submitter.innerHTML.indexOf("ADD METRIC") > -1) {
    addMetric(e);
  } else if (e.submitter.innerHTML.indexOf("ADD PART") > -1) {
    addPartLabel(e);
  } else {
    //this is the final SAVE function which does send the bottle back to cli after creating a JSON object with all the data.
    var bottleObject = {};
    bottleObject["kind"] = "Bottle";
    bottleObject["apiVersion"] = "data.act3-ace.io/v1";
    bottleObject["description"] = descriptionText.value;
    bottleObject["authors"] = [];
    for (var r = 0; r < authorTable.childElementCount; r++) {
      var author = {};
      author["name"] = authorTable.children[r].children[0].innerHTML;
      author["email"] = authorTable.children[r].children[1].innerHTML;
      author["url"] = authorTable.children[r].children[2].innerHTML;
      bottleObject["authors"].push(author);
    }
    var label = {};
    for (var l = 0; l <= labelList.childElementCount - 1; l++) {
      var text = labelList.children[l].innerHTML
        .substring(0, labelList.children[l].innerHTML.indexOf("<i"))
        .trim();
      var key =
        text.indexOf("=") > -1 ? text.substring(0, text.indexOf("=")) : text;
      var value =
        text.indexOf("=") > -1 ? text.substring(text.indexOf("=") + 1) : "";
      label[key] = value;
    }
    bottleObject["labels"] = label;
    bottleObject["metrics"] = [];
    for (var r = 0; r < metriclist.childElementCount; r++) {
      var metric = {};
      var text = metriclist.children[r].innerHTML
        .substring(0, metriclist.children[r].innerHTML.indexOf("<i"))
        .trim();
      var tooltip = metriclist.children[r].children[1];
      var key = text.substring(0, text.indexOf("="));
      var value = text.substring(text.indexOf("=") + 1);
      metric["name"] = key;
      metric["value"] = value;
      metric["description"] = tooltip.innerHTML;
      bottleObject["metrics"].push(metric);
    }
    bottleObject["parts"] = [];
    for (var r = 0; r < partsTable.childElementCount; r = r + 2) {
      var part = {};
      part["name"] = partsTable.children[r].children[1].innerHTML;
      part["size"] = parseInt(partsTable.children[r].children[2].innerHTML);
      part["digest"] = partsTable.children[r].children[0].innerHTML;
      //labels for parts
      part["labels"] = {};
      for (
        var pl = 0;
        pl <
        partsTable.children[r + 1].children[0].children[0].childElementCount;
        pl++
      ) {
        var text = partsTable.children[r + 1].children[0].children[0].children[
          pl
        ].innerHTML
          .substring(
            0,
            partsTable.children[r + 1].children[0].children[0].children[
              pl
            ].innerHTML.indexOf("<i")
          )
          .trim();
        var key =
          text.indexOf("=") > -1 ? text.substring(0, text.indexOf("=")) : text;
        var value =
          text.indexOf("=") > -1 ? text.substring(text.indexOf("=") + 1) : "";
        part["labels"][key] = value;
      }
      bottleObject["parts"].push(part);
    }
    bottleObject["sources"] = [];
    for (var r = 0; r < sourceTable.childElementCount; r++) {
      var author = {};
      author["name"] = sourceTable.children[r].children[0].innerHTML;
      author["uri"] = sourceTable.children[r].children[1].innerHTML;
      bottleObject["sources"].push(author);
    }

    anno = {};
    for (var l = 0; l <= annotationList.childElementCount - 1; l++) {
      var anno = {};
      var text = annotationList.children[l].innerHTML
        .substring(0, annotationList.children[l].innerHTML.indexOf("<i"))
        .trim();
      var tooltip = annotationList.children[l].children[1];
      anno[text] = tooltip.innerHTML;
    }
    bottleObject["annotations"] = anno;

    var xhr = new XMLHttpRequest();
    //open the request
    xhr.open("POST", window.location);
    xhr.setRequestHeader("Content-Type", "application/json");

    //send the form data
    xhr.send(JSON.stringify(bottleObject));

    xhr.onreadystatechange = function () {
      if (xhr.readyState == XMLHttpRequest.DONE) {
        if (xhr.status == 200) {
          isFormDirty = false;
          close();
        } else {
          submitErrorDiv.classList.remove("hide")
        }
      }
    };
  }
}

// Notify the server that we are discarding changes and close the window
function onDiscard(e) {
  var xhr = new XMLHttpRequest();
  //open the request
  xhr.open("POST", "/discard");
  xhr.send();
  xhr.onload = () => {
    close();
  };
}

// We use the Beacon API to send an async POST to /discard if the window is closed
// This will cause the server to shutdown
document.addEventListener('visibilitychange', function shutdown() {
  if (document.visibilityState === 'hidden') {
    navigator.sendBeacon('/discard');
  }
});


///AUTHORS
//adds a new author to the current table in the DOM
function addAuthor(e) {
  var validName = validateInput(authorName, true, 256);
  var validEmail = validateInput(
    authorEmail,
    true,
    256,
    /^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$/
  );
  if (validName && validEmail) {
    var row = authorTable.insertRow(authorTable.childElementCount);
    var cell1 = row.insertCell(0);
    cell1.innerHTML = authorName.value;
    var cell2 = row.insertCell(1);
    cell2.innerHTML = authorEmail.value;
    var cell3 = row.insertCell(2);
    cell3.innerHTML = authorUrl.value;
    var cell4 = row.insertCell(3);
    cell4.innerHTML =
      '<i class="pencil-icon invert-icon" onclick="editAuthor(' +
      (authorTable.childElementCount - 1).toString() +
      ');"></i><i class="close-icon invert-icon" onclick="deleteAuthor(' +
      (authorTable.childElementCount - 1).toString() +
      ')"></i>';
    authorName.value = "";
    authorEmail.value = "";
    authorUrl.value = "";
  }
}
//saving an author after editing
function saveAuthor(e) {
  var validName = validateInput(authorName, true, 256);
  var validEmail = validateInput(
    authorEmail,
    true,
    256,
    /^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$/
  );
  if (validName && validEmail) {
    var rowNum = 0;
    for (var r = 0; r < authorTable.childElementCount; r++) {
      if (authorTable.children[r].children[3].childElementCount === 1) {
        rowNum = r;
        break;
      }
    }

    authorTable.children[rowNum].children[0].innerHTML = authorName.value;
    authorTable.children[rowNum].children[1].innerHTML = authorEmail.value;
    authorTable.children[rowNum].children[2].innerHTML = authorUrl.value;
    authorTable.children[rowNum].children[3].innerHTML =
      '<i class="pencil-icon invert-icon" onclick="editAuthor(' +
      rowNum +
      ');"></i><i class="close-icon invert-icon" onclick="deleteAuthor(' +
      rowNum +
      ')"></i>';

    authorName.value = "";
    authorEmail.value = "";
    authorUrl.value = "";
    document.getElementById("authorAdd").innerHTML =
      '<i class="plus-icon"></i> ADD AUTHOR';
  }
}
//saving an author after editing
function updateAuthor(row) {

  var authorNameValue= document.getElementById('authornamevalue');
  var authorEmailValue = document.getElementById('authoremailvalue');
  var authorUrlValue = document.getElementById('authorurlvalue');

  var validName = validateInput(authorNameValue, true, 256);
  var validEmail = validateInput(
    authorEmailValue,
    true,
    256,
    /^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$/
  );
  if (validName && validEmail) {

    authorTable.children[row].classList.remove("edit-row");
    authorTable.children[row].children[0].innerHTML = authorNameValue.value;
    authorTable.children[row].children[1].innerHTML = authorEmailValue.value;
    authorTable.children[row].children[2].innerHTML = authorUrlValue.value;
    authorTable.children[row].children[3].innerHTML =
      '<i class="pencil-icon invert-icon" onclick="editAuthor(' +
      row +
      ');"></i><i class="close-icon invert-icon" onclick="deleteAuthor(' +
      row +
      ')"></i>';

  isFormDirty=true;    
}
}
//Edit Author
//The  code below allows the user to edit Author inputs(Name or email or url) inline
function editAuthor(row) {
  var authorNameValue= authorTable.children[row].children[0].innerHTML;
  var authorEmailValue = authorTable.children[row].children[1].innerHTML;
  var authorUrlValue = authorTable.children[row].children[2].innerHTML;
  //var authorButtonsValue = authorTable.children[row].children[3].innerHTML;
  authorTable.children[row].classList.add("edit-row");
  authorTable.children[row].children[0].innerHTML = '<input id="authornamevalue" type="text" value="'+ authorNameValue +' "/>';
  authorTable.children[row].children[1].innerHTML = '<input id="authoremailvalue" type="text" value="'+ authorEmailValue +'"/>';
  authorTable.children[row].children[2].innerHTML = '<input id="authorurlvalue" type="text" value="'+ authorUrlValue +'"/>';
  
  ///The line below defines a Save Author button and calls the saveAuthor function
  authorTable.children[row].children[3].innerHTML = 
  '<i class="save-icon" onclick="updateAuthor(' + 
  row +
  ')"/></i><i class="close-icon invert-icon" onclick="deleteAuthorUpdate(' +
  row +
  ')"></i>'
}

function deleteAuthorUpdate(row){
  authorTable.removeChild(authorTable.children[row]);
}

//deletes an author from the table
function deleteAuthor(row) {
  authorTable.removeChild(authorTable.children[row]);
  for (var r = 0; r < authorTable.childElementCount; r++) {
    authorTable.children[r].children[3].innerHTML =
      '<i class="pencil-icon invert-icon" onclick="editAuthor(' +
      r +
      ');"></i><i class="close-icon invert-icon" onclick="deleteAuthor(' +
      r +
      ')"></i>';
  }
}


//deletes an author from the table
function deleteAuthor(row) {
  authorTable.removeChild(authorTable.children[row]);
  for (var r = 0; r < authorTable.childElementCount; r++) {
    authorTable.children[r].children[3].innerHTML =
      '<i class="pencil-icon invert-icon" onclick="editAuthor(' +
      r +
      ');"></i><i class="close-icon invert-icon" onclick="deleteAuthor(' +
      r +
      ')"></i>';
  }
}


///SOURCES
//Allows user to edit Source inputs inline
function editSource(row) {
  sourceNameValue = sourceTable.children[row].children[0].innerHTML;
  sourceUriValue = sourceTable.children[row].children[1].innerHTML;
  sourceTable.children[row].children[2].removeChild(
    sourceTable.children[row].children[2].children[0]
  );
  sourceTable.children[row].classList.add("edit-row");
  sourceTable.children[row].children[0].innerHTML = '<input id="sourcenamevalue" type="text" value="'+ sourceNameValue +' "/>';
  sourceTable.children[row].children[1].innerHTML = '<input id="sourceurivalue" type="text" value="'+ sourceUriValue +'"/>';
  sourceTable.children[row].children[2].innerHTML = 
  '<i class="save-icon" onclick="updateSource(' +
   row + 
   ')"/></i><i class="close-icon invert-icon" onclick="deleteSourceUpdate(' +
  row +
  ')"></i>';
  

  document.getElementById("sourceAdd").innerHTML =
    '<i class="pencil-icon"></i> SAVE SOURCE';
}

function deleteSourceUpdate(row){
  sourceTable.removeChild(sourceTable.children[row]);
}
//deletes the selected table row from the DOM
function deleteSource(row) {
  sourceTable.removeChild(sourceTable.children[row]);
  for (var r = 0; r < sourceTable.childElementCount; r++) {
    sourceTable.children[r].children[2].innerHTML =
      '<i class="pencil-icon invert-icon" onclick="editSource(' +
      r +
      ');"></i><i class="close-icon invert-icon" onclick="deleteSource(' +
      r +
      ')"></i>';
  }
}
//adds a new source to the current table in the DOM
function addSource(e) {
  var validName = validateInput(sourceName, true, 256);
  var validURI = validateInput(sourceUrl, true);
  if (validName && validURI) {
    var row = sourceTable.insertRow(sourceTable.childElementCount);
    var cell1 = row.insertCell(0);
    cell1.innerHTML = sourceName.value;
    var cell2 = row.insertCell(1);
    cell2.innerHTML = sourceUrl.value;
    var cell4 = row.insertCell(2);
    cell4.innerHTML =
      '<i class="pencil-icon invert-icon" onclick="editSource(' +
      (sourceTable.childElementCount - 1).toString() +
      ');"></i><i class="close-icon invert-icon" onclick="deleteSource(' +
      (sourceTable.childElementCount - 1).toString() +
      ')"></i>';
    sourceName.value = "";
    sourceUrl.value = "";
  }
}
//saves an annotation after editing
function saveSource(e) {
  var validName = validateInput(sourceName, true, 256);
  var validURI = validateInput(sourceUrl, true);
  if (validName && validURI) {
    var rowNum = 0;
    for (var r = 0; r < sourceTable.childElementCount; r++) {
      if (sourceTable.children[r].children[2].childElementCount === 1) {
        rowNum = r;
        break;
      }
    }
    sourceTable.children[row].classList.add("edit-row");
    sourceTable.children[rowNum].children[0].innerHTML = sourceName.value;
    sourceTable.children[rowNum].children[1].innerHTML = sourceUrl.value;
    sourceTable.children[rowNum].children[2].innerHTML =
      '<i class="pencil-icon invert-icon" onclick="editSource(' +
      rowNum +
      ');"></i><i class="close-icon invert-icon" onclick="deleteSource(' +
      rowNum +
      ')"></i>';

    sourceName.value = "";
    sourceUrl.value = "";
    document.getElementById("sourceAdd").innerHTML =
      '<i class="plus-icon"></i> ADD SOURCE';
  }
}

function updateSource(rowNum) {
  var sourceNameValue = document.getElementById('sourcenamevalue');
  var sourceUriValue = document.getElementById('sourceurivalue');

  var validName = validateInput(sourceNameValue, true, 256);
  var validURI = validateInput(sourceUriValue, true);

  if (validName && validURI) {
    sourceTable.children[rowNum].children[0].innerHTML = sourceNameValue.value;
    sourceTable.children[rowNum].children[1].innerHTML = sourceUriValue.value;
    sourceTable.children[rowNum].children[2].innerHTML =
      '<i class="pencil-icon invert-icon" onclick="editSource(' +
      rowNum +
      ');"></i><i class="close-icon invert-icon" onclick="deleteSource(' +
      rowNum +
      ')"></i>';

    isFormDirty = true;
}
}
///LABELS
//removes a label from the list in the DOM
function deleteLabel(key) {
  labelList.removeChild(labelList.children[getIndex(labelList, key)]);
}
//Adds a label to the list in the DOM
function saveLabel(e) {
  var validName = validateInput(
    labelName,
    true,
    63,
    /^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$/
  );
  var validValue = validateInput(
    labelValue,
    false,
    63,
    /(?![0-9]+$)(?!.*-$)(?!-)[a-zA-Z0-9-]{1,63}$/gm
  );
  if (validName && validValue) {
    var li = document.createElement("li");
    li.innerHTML =
      labelName.value +
      (labelValue.value ? "=" + labelValue.value : "") +
      '<i class="close-icon" onclick="deleteLabel(\'' +
      labelName.value +
      "');\"></i>";
    labelList.appendChild(li);
    labelName.value = "";
    labelValue.value = "";
  }
}
///ANNOTATIONS
//removes an annotation from the list in the DOM
function deleteAnnotations(key) {
  annotationList.removeChild(annotationList.children[getIndex(annotationList, key)]);
}
//adds an annotation from the list in the DOM
function saveAnnotation(e) {
  var validName = validateInput(
    annotationName,
    true,
    63,
    /^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$/
  );
  var validValue = validateInput(
    annotationValue,
    true,
    0
  );
  if (validName && validValue) {
    var li = document.createElement("li");
    li.innerHTML =
      annotationName.value +
      '<i class="close-icon" onclick="deleteAnnotations(\'' +
      annotationName.value +
      '\');"></i><span class="tooltiptext">' +
      annotationValue.value +
      "</span>";
    li.classList.add("tooltip");
    annotationList.appendChild(li);
    annotationName.value = "";
    annotationValue.value = "";
  }
}

///Metrics
//removes a metric from the list in the DOM
function deleteMetric(key) {
  metriclist.removeChild(metriclist.children[getIndex(metriclist, key)]);
}

//adds a metric to the list in the DOM
function addMetric(e) {
  var validName = validateInput(
    metricname,
    true,
    63,
    /^(([A-Za-z0-9][-A-Za-z0-9_. ]*)?[A-Za-z0-9])?$/
  );
  var validValue = validateInput(
    metricvalue,
    true,
    63,
    /^-?\d*(\.\d+)?$/g
  );
  if (validName && validValue) {
    var li = document.createElement("li");
    li.innerHTML =
      metricname.value +
      "=" +
      metricvalue.value +
      '<i class="close-icon" onclick="deleteMetric(\'' +
      metricname.value +
      '\');"></i><span class="tooltiptext">' +
      metricDesc.value +
      "</span>";
    li.classList.add("tooltip");
    metriclist.appendChild(li);
    metricname.value = "";
    metricvalue.value = "";
    metricDesc.value = "";
  }
}

///PART LABELS
//removes a part label from the list in table for a given part the DOM
function deletePartLabel(part, label) {
  var partLabelList = partsTable.children[part + 1].children[0].children[0];
  partLabelList.removeChild(partLabelList.children[getIndex(partLabelList, label)]);
}


//adds a part label to the list in table for the selected part the DOM
function addPartLabel(e) {
  var validName = validateInput(
    partName,
    true,
    63,
    /^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$/
  );
  var validValue = validateInput(
    partValue,
    false,
    63,
    /^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$/
  );
  var selectAll = document.getElementById('part_checked_All')
  if (validName && validValue) {
    for (var p = 0; p < partsTable.childElementCount; p = p + 2) {
      if (partsTable.children[p].children[0].children[0].checked || selectAll.checked) {
        var partLabelList = partsTable.children[p + 1].children[0].children[0];
        var li = document.createElement("li");

        li.innerHTML =
          partName.value +
          "=" +
          partValue.value +
          '<i class="close-icon" onclick="deletePartLabel(' +
          p.toString() +
          ",'" +
           partName.value +
          '\');"></i>';
        partLabelList.append(li);
      }
    }
    partName.value = "";
    partValue.value = "";
  }
}

///UTILITIES
//validates form inputs before submission
function validateInput(element, required, maxLength, regEx) {
  console.log(element)
  //clear existing validation
  if (element.classList.contains("invalid")) {
    element.classList.remove("invalid");
    element.parentNode.removeChild(element.nextSibling);
  }
  //check
  if (required && element.value.length == 0) {
    addError(element, "This is required");
    return false;
  } else if (
    element.value &&
    maxLength > 0 &&
    element.value.length > maxLength
  ) {
    addError(element, "Input must be less/equal to " + maxLength.toString());
    return false;
  } else if (element.value && regEx && !regEx.test(element.value)) {
    addError(element, "Input must be a valid format.");
    return false;
  } else {
    return true;
  }
}

//adds error messages below the elements if input's content is invalid
function addError(element, message) {
  element.classList.add("invalid");
  var wrapper = document.createElement("div");
  wrapper.classList.add("error-element");
  // insert wrapper before el in the DOM tree
  if (!element.parentNode.classList.contains("error-element")) {
    element.parentNode.insertBefore(wrapper, element);
    wrapper.appendChild(element);
  } else {
    wrapper = element.parentNode;
  }
  // move el into wrapper
  var errorMsg = document.createElement("p");
  errorMsg.innerHTML = message;
  errorMsg.classList.add("error");
  wrapper.appendChild(errorMsg);
}


//due to Go's template not providing a key for an object's index in object loop
//helper function to convert a key to row index, and return it to the function
function getIndex(list, key) {
  for (var r = 0; r < list.childElementCount; r++) {
    var text = list.children[r].innerHTML
      .substring(0, list.children[r].innerHTML.indexOf("<i"))
      .trim();
    var pillKey =
      text.indexOf("=") > -1 ? text.substring(0, text.indexOf("=")) : text;
    if (pillKey === key) {
      return r;
    }
  }
}
