import sys
from fpdf import FPDF

dict_word={}
scenario_sleep = False
scenario_drop = False
# reading entire line from STDIN (standard input)
for line in sys.stdin:
    
    # to remove leading and trailing whitespace
    line = line.strip()
    if "sleepDuration" in line:        
        words = line.split()
        for val in words:
            if "sleepDuration" in val:
                st_word=val.split(":")
                scenario_time=st_word[-1][:-1]
        if scenario_time != "0":
            scenario_sleep=True
            break



# Python program to create
# a pdf file

# save FPDF() class into a variable pdf
pdf = FPDF()

# Add a page
pdf.add_page()
pdf.set_font("Times",'BU', size = 25)
pdf.cell(200, 10, txt = "Results ",align = 'C',ln=1)

# set style and size of font
# that you want in the pdf
pdf.set_font("Times",'B',size = 17)


if scenario_sleep:  
    # create a cell
    pdf.cell(200, 10, txt = "Scenario Detected: Sleep",ln = 3)
    pdf.set_font("Times", size = 10)
    # add another cell
    pdf.cell(200, 10, txt = "Inference from logs: A sleep scenario has been detected. The chaos issuer has decided to sleep for "+ scenario_time+"s." ,ln = 8)
    pdf.cell(200, 10, txt = "Repercussions: This is affecting the renewal of certificate and could cause the application to communicate  over insecure channels." ,ln = 8)
    pdf.cell(200, 10, txt = "-------------------------------------------------------------------------------------------------------------------------------------------------------------------" ,ln = 8)
    # save the pdf with name .pdf
    pdf.output("result.pdf")








